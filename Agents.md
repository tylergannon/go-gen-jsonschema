Run all Go unit tests before doing anything, to establish a baseline. Your job
CANNOT be considered complete if you do not run tests after completing the work.
Therefore it is imperative to verify that tests are working before you begin.

If there is an issue running tests at all (i.e. missing modules that can't be loaded),
STOP.
If there are broken tests, begin by fixing those broken tests.

DO NOT consider the job to be complete until unit tests have been updated
and `go test ./...` completes successfully.

If you have been asked to add new functionality, you must write unit tests
that verify the new functionality.
If you have been asked to perform a refactoring, you need only change unit
tests to the extent needed to ensure all functionality has good tests.
