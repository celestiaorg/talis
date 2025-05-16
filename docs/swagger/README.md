# Swagger Documentation

This directory contains the Swagger/OpenAPI documentation for the Talis API. This README provides information about how Swagger is implemented in this Go project, how to update the documentation, and how to use it effectively.

## What is Swagger/OpenAPI?

[Swagger](https://swagger.io/) (now known as OpenAPI) is a specification for machine-readable interface files for describing, producing, consuming, and visualizing RESTful web services. It provides a standardized way to document APIs, making it easier for developers to understand and interact with them.

In this project, we use [Swaggo](https://github.com/swaggo/swag) to generate Swagger documentation from annotations in our Go code.

## Files in this Directory

- **docs.go**: Contains the generated Go code that registers the Swagger documentation with the application.
- **swagger.json**: The Swagger specification in JSON format, generated from the annotations in the Go code.
- **swagger.yaml**: The Swagger specification in YAML format, equivalent to swagger.json but in a different format.
- **swagger-ui.html**: The HTML file that serves the Swagger UI, providing an interactive interface for exploring the API.

## How to Update the Documentation

The Swagger documentation is generated from annotations in the Go code. To update the documentation:

1. Add or modify the Swagger annotations in your Go code. These annotations are comments that start with `@` and provide information about your API endpoints, request parameters, response types, etc.

2. Run the following command to regenerate the Swagger documentation:

   ```bash
   make swagger
   ```

   This command runs `swag init` with the appropriate parameters to generate the documentation files in this directory.

3. Verify the changes by accessing the Swagger UI at `/swagger` when the application is running.

## Swagger Annotations

Swagger annotations are special comments in your Go code that provide information about your API. Here are some common annotations:

### General API Information

```go
// @title Talis API
// @version 1.0
// @description API for Talis - Celestia's infrastructure management service
// @host localhost:8080
// @BasePath /api/v1
```

Note: The `@host` value can be overridden at runtime by setting the `API_HOST` environment variable. This allows you to dynamically set the host based on the deployment environment (e.g., `localhost:8080` for local development, `api.example.com` for production).

### Endpoint Documentation

```go
// @Summary Get instance details
// @Description Returns detailed information about a specific instance identified by its ID.
// @Tags instances
// @Accept json
// @Produce json
// @Param id path int true "Instance ID" example(123)
// @Success 200 {object} models.Instance "Complete instance details"
// @Failure 400 {object} types.ErrorResponse "Invalid input"
// @Failure 500 {object} types.ErrorResponse "Internal server error"
// @Router /instances/{id} [get]
```

### Model Documentation

```go
// swagger:model
// Example: {"id":123,"status":"running"}
type Response struct {
    // Unique identifier for the infrastructure job
    ID uint `json:"id"`
    
    // Current status of the infrastructure job
    Status string `json:"status"`
}
```

For a complete reference of Swagger annotations, see the [Swaggo documentation](https://github.com/swaggo/swag#declarative-comments-format).

## Accessing the Swagger UI

When the application is running, you can access the Swagger UI at:

```
http://localhost:8080/swagger
```

This provides an interactive interface where you can:

- Browse all available API endpoints
- See detailed information about request parameters and response types
- Try out API calls directly from the browser
- View example requests and responses


## Best Practices

1. **Keep annotations up to date**: Always update the Swagger annotations when you modify an API endpoint.
2. **Include examples**: Use the `example()` annotation to provide example values for parameters.
3. **Provide detailed descriptions**: Write clear and concise descriptions for endpoints and parameters.
4. **Document all responses**: Include documentation for both success and error responses.
5. **Group related endpoints**: Use the `@Tags` annotation to group related endpoints together.

## Useful Resources

- [Swaggo GitHub Repository](https://github.com/swaggo/swag)
- [Swagger/OpenAPI Specification](https://swagger.io/specification/)
- [Swagger UI](https://swagger.io/tools/swagger-ui/)
- [OpenAPI Initiative](https://www.openapis.org/)
- [Swagger Editor](https://editor.swagger.io/)
- [Swagger Article](https://medium.com/julotech/implementing-swagger-in-go-projects-8579a5fb955)

## Troubleshooting

### Common Issues

1. **Swagger documentation not updating**: Make sure you're running `make swagger` after making changes to the annotations.
2. **Missing endpoints in Swagger UI**: Check that your endpoints have the correct Swagger annotations.
3. **Incorrect parameter types**: Verify that the parameter types in your annotations match the actual types in your code.
4. **Host not updating dynamically**: If the Swagger UI is not showing the correct host after setting the `API_HOST` environment variable, ensure that the application was restarted after changing the environment variable and that the variable is correctly set in your environment or `.env` file.

If you encounter any issues with the Swagger documentation, please check the Swaggo documentation or open an issue in the project repository.
