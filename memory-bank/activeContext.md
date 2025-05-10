# Active Context: stui

## Current Focus Areas

1. **Sorting Functionality**:
   - Implementing Ctrl+s shortcut for sorting
   - Developing sort modal interface (sortmodal.go)
   - Building sort selector components (selector_sort.go)

2. **Command Modal Improvements**:
   - Enhancing command execution flow
   - Improving batch operation handling
   - Refactoring command definitions (commandmodal.go)

3. **View Components**:
   - Main application view refinements (app.go)
   - StuiView component updates (stuiview.go)

## Recent Changes

- Added initial sorting modal components
- Refactored command modal architecture
- Improved view rendering performance

## Key Decisions

1. **Sorting Implementation**:
   - Using modal dialog for sort configuration
   - Supporting multiple sort criteria
   - Maintaining sort state between refreshes

2. **Command Execution**:
   - Centralized command registry
   - Context-aware command availability
   - Batch operation support

## Next Steps

1. Complete sorting functionality:
   - Implement sort application to views
   - Add persistent sort preferences
   - Document sort usage

2. Refactor command system:
   - Reduce duplication in command definitions
   - Improve command discovery
   - Add plugin support framework

3. Address technical debt:
   - Connection state tracking
   - Selected row highlighting
   - Error handling improvements
