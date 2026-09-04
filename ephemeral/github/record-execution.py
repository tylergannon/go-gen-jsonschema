import json, subprocess
from pathlib import Path
repo='tylergannon/go-gen-jsonschema'
out=Path('ephemeral/github')
def gh(*args):
    return subprocess.check_output(['gh',*args],text=True)
def api(path): return json.loads(gh('api',f'repos/{repo}/{path}'))
m=api('milestones/1')
plan='''\n\nExecution order\n\n1. Contract gate: #56. In parallel, fix #33/#34 and the unambiguous named-scalar, property-escaping, and deterministic-order defects in #59. Strict/empty-enum policy waits for #56.\n2. Three parallel implementation tracks after the contract: codecs #57 then #58; agent CLI #60 then #61; helpers/cleanup #59 and #63. Keep one owner for the shared containing-struct codec architecture.\n3. Complete product conformance #62 and plugin #64 in parallel against the merged foundations.\n4. Complete real-agent evaluations #65, fix discovered product/workflow failures, then perform exact-candidate release proof and publication #66.\n\nNative issue dependencies express completion gates, not a ban on preparatory work. Start #62 conformance cases and #65 tasks/graders after #56, plugin packaging during CLI work, and #66 licensing/platform decisions early. Run an early end-to-end union workflow through preview, generation, schema-valid roundtrip, and clean regeneration.\n\nUse separate worktrees for implementation tracks and integrate tested changes through reviewable PRs. #36 is already fixed by PR #38; an initially empty module-cache full-suite run on 2026-09-04 passed. App-server and streaming remain beyond 1.0.'''
payload={'description':m['description'].split('\n\nExecution order')[0]+plan}
p=out/'milestone-execution.json';p.write_text(json.dumps(payload))
gh('api',f'repos/{repo}/milestones/1','--method','PATCH','--input',str(p))
edges={57:[56],58:[56,57,33],59:[56],60:[56,34],61:[60],62:[56,57,58,59,33],63:[56],64:[60,61,63,57,58],65:[64,62],66:[62,63,65]}
for n, blockers in edges.items():
    current={x['number'] for x in api(f'issues/{n}/dependencies/blocked_by')}
    missing=[b for b in blockers if b not in current]
    if missing: gh('issue','edit',str(n),'--repo',repo,'--add-blocked-by',','.join(map(str,missing)))
    actual={x['number'] for x in api(f'issues/{n}/dependencies/blocked_by')}
    assert set(blockers)<=actual,(n,blockers,actual)
    print(f'#{n} blocked by {sorted(actual)}',flush=True)
# Align written prerequisites with the native graph and completed cold-cache proof.
for n in [58,59,60,66]:
    i=api(f'issues/{n}');body=i['body']
    if n==58: body=body.replace('Depends on #56. Related: #33, #46.', 'Depends on #56, #57, and #33. #57 establishes the shared containing-struct codec architecture; enum fixture/design work can proceed in parallel. Related: #46.')
    if n==59: body=body.replace('Coordinate strict/error policy with #56.', 'Completion depends on #56 for strict/error policy. Named-scalar, property-escaping, and deterministic-order fixes may proceed in parallel with that contract work.')
    if n==60: body=body.replace('Depends on #56. Related: #34.', 'Depends on #56 and #34.')
    if n==66:
        body=body.replace('- Close the cold-cache proof gap in #36. Its specific go mod tidy stderr assertion is already fixed on main; verify the remaining acceptance rather than duplicate that fix.', '- Cold-cache prerequisite #36 is satisfied: PR #38 fixed the stderr assertion; an initially empty module-cache `go test -count=1 -v ./...` run passed on 2026-09-04. Retain clean-cache coverage in release verification.')
        body=body.replace('Depends on #62, #63, #65, #33, #34, and #36.', 'Depends on #62, #63, and #65; #33/#34 are prerequisites through those dependencies. #36 is already satisfied.')
    p=out/f'issue-{n}.md';p.write_text(body)
    gh('issue','edit',str(n),'--repo',repo,'--body-file',str(p))
verified={str(n):[x['number'] for x in api(f'issues/{n}/dependencies/blocked_by')] for n in edges}
(out/'verified-dependencies.json').write_text(json.dumps(verified,indent=2)+'\n')
(out/'verified-milestone.json').write_text(json.dumps(api('milestones/1'),indent=2)+'\n')
