# Git Conventions

## Autonomous Commits

Commit changes autonomously when it makes sense — after completing a logical unit of work, after fixing a bug, or after finishing a feature. Don't wait for the user to ask.

Good commit points:
- After completing a service implementation
- After finishing a group of related changes (e.g., all proto restructuring)
- After fixing build errors or resolving issues
- After adding a new feature or component

## Commit Messages

Concise, imperative mood. Focus on "why" not "what".

```
Add gateway GraphQL server with mock data
Restructure proto files into pkg/<service>/
Wire up builder gRPC server with stub implementation
```

## Branch Strategy

Work on `main` for now. Feature branches when the team grows.
