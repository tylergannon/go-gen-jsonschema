# Generated Tests

After generating structs and unmarshaling functions, we generate some JSON
samples of the structured responses, as well as unit tests that validate that
each of them.

1. Generate N values.  On the ith sample, for any given union type containing m
   alternatives, choose the i%m-th option.  Meaning, on the 7th sample, for a
   union type containing five alternatives, choose index `6 % 5 -> 1` and
   instantiate that option.
2. Validate sample data before storing it.  If there are any failures, report this
   after generation, as this means that there's a high likelihood the structured
   response schema should be revisited.
3. Store the results in a "testfixtures/jsonschema" directory.
4. For each sample, write a test case.
5. Send the flattened struct type to the LLM along with the sample, to make test
   cases.
6. Offer a tool function `ResolveInterfaceInstance`.

   Parameters:
     - Interface name
     - Discriminator value

   Response:
   - The package alias and type name for the struct type expected on the given interface/discriminator
   - The flattened struct type
7. For each interface field on the struct, there should be a type assertion followed by
   a set of assertions for the individual fields.
8. Yes, #7 should be recursive.

