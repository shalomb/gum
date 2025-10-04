# Gum Backlog

## High Priority

### Repository Metadata Sync
- **Status**: In Progress
- **Description**: Daily sync of GitHub repository metadata (names, orgs, topics, timestamps)
- **Implementation**: GitHub API integration with daily crontab updates
- **Benefits**: Rich project signposts, automatic metadata maintenance

## Medium Priority

### GitHub Codespaces Discovery
- **Status**: Backlog
- **Description**: Discover and track GitHub Codespaces for development environment awareness
- **Requirements**: 
  - `codespace` scope for user codespaces
  - `admin:org` scope for organization codespaces
  - Privacy considerations for sensitive data
- **Use Cases**:
  - Track active development environments
  - Team collaboration insights
  - Project activity analysis
  - Resource optimization
- **Implementation Notes**:
  - Optional scope (graceful fallback if unavailable)
  - Real-time data with appropriate caching
  - Privacy-first approach
  - Rate limiting considerations

### Library Architecture
- **Status**: Backlog
- **Description**: Refactor gum into a library for use by other applications
- **Benefits**:
  - Shared cache across services
  - Reusable project discovery logic
  - Rich metadata access for external tools
  - Performance benefits through shared resources

### Advanced Index Generation
- **Status**: Backlog
- **Description**: Generate comprehensive INDEX.md style reports with rich metadata
- **Features**:
  - Smart project classification
  - Topic-based categorization
  - Activity-based filtering
  - Export to multiple formats (Markdown, JSON, CSV)

## Low Priority

### Plugin System
- **Status**: Backlog
- **Description**: Extensible plugin system for custom analyzers and integrations
- **Use Cases**:
  - Custom project type detection
  - Integration with external tools
  - Custom metadata extraction
  - Organization-specific workflows

### Advanced Search and Filtering
- **Status**: Backlog
- **Description**: Enhanced search capabilities with complex queries
- **Features**:
  - Boolean queries (AND, OR, NOT)
  - Date range filtering
  - Numeric range filtering
  - Saved searches and filters

### Analytics Dashboard
- **Status**: Backlog
- **Description**: Web-based dashboard for project analytics and insights
- **Features**:
  - Project health metrics
  - Team activity visualization
  - Technology stack analysis
  - Trend analysis over time

## Completed

### Version Command
- **Status**: Completed
- **Description**: Added version subcommand with build information
- **Features**: Git commit, build date, Go version, OS/Arch info, verbose mode

### Locate Integration
- **Status**: Completed
- **Description**: Fast project discovery using system locate database
- **Performance**: 34x speed improvement for large directories
- **Features**: Automatic detection, smart fallback, database monitoring

### Frecency Algorithm
- **Status**: Completed
- **Description**: Intelligent directory ranking combining frequency and recency
- **Features**: Logarithmic frequency scaling, multi-tier exponential decay