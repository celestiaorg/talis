# Talis Integration Test Package


## Features

- **TestSuite**: Complete test setup including:
  - In-memory database
  - Real API server
  - Real API client
  - Mocked external providers

- **Mock Management**: 
  - Centralized mock implementations
  - Standardized configurations
  - Provider factory functions (TODO)

## Testing Architecture

### Package Level Unit Test

Package level testing should be unit level testing. Tests should not need any context outside of the package itself in order to be executed. With the exception of API packages, little to no mocking should be required for the unit tests. 

### API Interactions

API interactions need to be mocked so that we are not using production billed endpoints in testing. 
There should be a single mock for an API package created in `test/mocks` with a set of standard successful and error responses predefined that tests can use. 
By defining a standard set of success and error responses, tests can focus on ensuring the function logic is responding to these successful and error API responses. 

### Integration Tests

Integration tests target the entire application from API to DB. API responses should us the standard mocked responses from `test/mocks` and the DB should be an in memory DB. 
Beside the API mocks and the in memory DB, everything else should be targeted directly from a testing perspective. 
