import assert from "node:assert/strict";
import { execFileSync, spawnSync } from "node:child_process";
import { createHash } from "node:crypto";
import { cpSync, existsSync, mkdirSync, mkdtempSync, readFileSync, readdirSync, rmSync, symlinkSync, writeFileSync } from "node:fs";
import { dirname, join, resolve } from "node:path";
import { fileURLToPath } from "node:url";

const here = dirname(fileURLToPath(import.meta.url));
const repo = resolve(here, "../..");
const proofRoot = resolve(process.env.TS_PROOF_DIR || join(repo, "ephemeral/typescript-generation/proof"));
mkdirSync(proofRoot, { recursive: true });
const run = mkdtempSync(join(proofRoot, "run-"));
const consumer = join(run, "consumer");
const cli = join(run, "polytype");
const logs = [];
const claims = [];
const go = execFileSync("go", ["env", "GOROOT"], { encoding: "utf8" }).trim() + "/bin/go";
const goVersion = execFileSync(go, ["env", "GOVERSION"], { encoding: "utf8" }).trim().replace(/^go/, "");
const ts = join(here, "node_modules/typescript/bin/tsc");
assert.ok(existsSync(ts), "Run npm ci in tests/typescript before the dedicated compiler lane");

function command(executable, args, options = {}, expected = 0) {
  const result = spawnSync(executable, args, { encoding: "utf8", ...options });
  const output = (result.stdout || "") + (result.stderr || "");
  logs.push(`$ ${executable} ${args.join(" ")}\nexit=${result.status}\n${output}`);
  if (expected === "failure") assert.notEqual(result.status, 0, output);
  else assert.equal(result.status, expected, output || String(result.error));
  return output;
}
function claim(name) { claims.push(name); process.stdout.write(`PASS ${name}\n`); }
function snapshot(path) {
  return Object.fromEntries(readdirSync(path).sort().map(name => [name,
    createHash("sha256").update(readFileSync(join(path, name))).digest("hex")]));
}
function compiler(files, expected = 0) {
  writeFileSync(join(consumer, "tsconfig.json"), JSON.stringify({
    compilerOptions: { strict: true, exactOptionalPropertyTypes: true,
      noUncheckedIndexedAccess: true, noEmit: true, target: "ES2022",
      module: "NodeNext", moduleResolution: "NodeNext", types: [] }, files,
  }, null, 2) + "\n");
  return command(process.execPath, [ts, "--project", join(consumer, "tsconfig.json"), "--pretty", "false"], {}, expected);
}

try {
  const projection = join(run, "projection");
  mkdirSync(projection);
  for (const file of ["consumer.ts", "empty-consumer.ts", "tsconfig.json"]) {
    cpSync(join(here, "projection", file), join(projection, file));
  }
  command(go, ["run", "./tests/typescript/projection/generate", join(projection, "generated")], { cwd: repo });
  command(process.execPath, [ts, "--project", join(projection, "tsconfig.json"), "--pretty", "false"]);
  claim("Fresh backend projection output passes exact-literal, empty-module, discriminator, Unicode, and escaping compiler obligations");
  command(go, ["build", "-o", cli, "./polytype"], { cwd: repo });
  mkdirSync(consumer);
  cpSync(join(here, "fixture"), consumer, { recursive: true });
  writeFileSync(join(consumer, "go.mod"), `module example.com/ts-consumer\n\ngo ${goVersion}\n\nrequire github.com/tylergannon/polytype v0.0.0\nreplace github.com/tylergannon/polytype => ${JSON.stringify(repo)}\n`);
  writeFileSync(join(consumer, "package.json"), '{"type":"module"}\n');
  command(go, ["mod", "tidy"], { cwd: consumer });
  const goOnlyPath = join(run, "go-only-path");
  mkdirSync(goOnlyPath);
  symlinkSync(go, join(goOnlyPath, "go"));
  const generationEnv = { ...process.env, PATH: goOnlyPath, CGO_ENABLED: "0", GOTOOLCHAIN: "local" };
  delete generationEnv.JSONSCHEMA_NO_CHANGES;
  assert.ok(spawnSync("node", ["--version"], { env: generationEnv }).error, "Node must be unavailable to the generator");
  const generate = (args = [], expected = 0) => command(cli, ["gen", "--target", consumer, ...args], { cwd: run, env: generationEnv }, expected);
  const generated = join(consumer, "generated");
  // Relative output is relative to invocation CWD, independently of --target.
  generate(["--typescript", "consumer/generated", "--typescript-barrel"]);
  assert.deepEqual(readdirSync(generated).sort(), ["index.ts", "types.ts"]);
  assert.match(readFileSync(join(generated, "index.ts"), "utf8"), /export type/);
  claim("CLI generates the requested relative directory and barrel with Node absent from PATH");
  const envelopeSchema = JSON.parse(readFileSync(join(consumer, "jsonschema/Envelope.json"), "utf8"));
  assert.deepEqual(envelopeSchema.properties.status.enum, ["ready", 'wait"ing', "converted"]);
  assert.deepEqual(envelopeSchema.properties.priority.enum, [0, 1, 8, 4]);
  assert.deepEqual(envelopeSchema.properties.priority_name.enum, ["Low", "High", "Urgent", "Medium"]);
  claim("JSON Schema and TypeScript share exact typed enum membership and evaluated values");

  const original = snapshot(generated);
  generate(["--typescript", generated, "--typescript-barrel"]);
  assert.deepEqual(snapshot(generated), original);
  generate(["--typescript", generated, "--typescript-barrel", "--no-changes"]);
  assert.deepEqual(snapshot(generated), original);
  claim("Absolute paths, deterministic regeneration, and current-output no-change checking work");

  const typesPath = join(generated, "types.ts");
  const types = readFileSync(typesPath, "utf8");
  writeFileSync(typesPath, types + "// deliberately stale\n");
  const stale = snapshot(generated);
  const staleError = generate(["--typescript", generated, "--typescript-barrel", "--no-changes"], "failure");
  assert.match(staleError, /typescript|types\.ts/i);
  assert.deepEqual(snapshot(generated), stale);
  generate(["--typescript", generated, "--typescript-barrel"]);
  assert.deepEqual(snapshot(generated), original);
  claim("Stale TS fails no-change checking without mutation even with unchanged schemas");

  generate(["--typescript", generated]);
  assert.ok(!existsSync(join(generated, "index.ts")));
  const withoutBarrel = snapshot(generated);
  generate(["--typescript", generated, "--typescript-barrel", "--no-changes"], "failure");
  assert.deepEqual(snapshot(generated), withoutBarrel);
  generate(["--typescript", generated, "--typescript-barrel"]);
  assert.deepEqual(snapshot(generated), original);
  claim("Barrel creation, omission, removal, and missing-barrel stale detection work");

  const collision = join(consumer, "user-owned");
  mkdirSync(collision);
  writeFileSync(join(collision, "types.ts"), "export type UserOwned = string;\n");
  const ownedByUser = snapshot(collision);
  generate(["--typescript", collision], "failure");
  assert.deepEqual(snapshot(collision), ownedByUser);
  claim("An existing user-owned output is preserved with a failing diagnostic");

  for (const file of ["consumer.ts", "missing-case.ts"]) cpSync(join(here, file), join(consumer, file));
  compiler(["consumer.ts"]);
  claim("Actual output passes strict TS positive, negative, narrowing, presence, enum, reference, and composition obligations");
  const incomplete = compiler(["missing-case.ts"], "failure");
  assert.match(incomplete, /TS2345/);
  assert.match(incomplete, /never/);
  writeFileSync(join(run, "missing-case-diagnostics.txt"), incomplete);
  claim("An incomplete discriminated-union switch fails exhaustiveness checking");

  const typesGoPath = join(consumer, "types.go");
  const typesGo = readFileSync(typesGoPath, "utf8");
  writeFileSync(typesGoPath, typesGo.replace("// ADDED_VARIANT", "func (Renamed) event() {}"));
  generate(["--typescript", generated, "--typescript-barrel"]);
  const added = compiler(["consumer.ts"], "failure");
  assert.match(added, /TS2345/);
  assert.match(added, /renamed|Renamed/);
  writeFileSync(join(run, "added-variant-diagnostics.txt"), added);
  claim("Adding a qualifying implementation of the sealed interface breaks a previously complete exhaustive handler");
  writeFileSync(typesGoPath, typesGo);
  generate(["--typescript", generated, "--typescript-barrel"]);
  assert.deepEqual(snapshot(generated), original);
  compiler(["consumer.ts"]);
  command(go, ["test", "./..."], { cwd: consumer });
  claim("The generated fresh Go consumer still builds and tests");
} finally {
  writeFileSync(join(run, "commands.log"), logs.join("\n"));
  writeFileSync(join(run, "proof.json"), JSON.stringify({
    head: execFileSync("git", ["rev-parse", "HEAD"], { cwd: repo, encoding: "utf8" }).trim(),
    workingTree: execFileSync("git", ["status", "--short"], { cwd: repo, encoding: "utf8" }).trim(),
    typescript: JSON.parse(readFileSync(join(here, "node_modules/typescript/package.json"), "utf8")).version,
    go: goVersion, claims,
  }, null, 2) + "\n");
  // The symlink is an operating aid, not a portable retained artifact.
  rmSync(join(run, "go-only-path"), { recursive: true, force: true });
  rmSync(cli, { force: true });
  process.stdout.write(`Proof retained at ${run}\n`);
}
