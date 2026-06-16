# openapi-trim

A lightweight tool to extract and download specific API modules from large OpenAPI/Swagger documentation. Instead of handling massive OpenAPI JSON files, this tool lets you select only the modules (tags) you need and download a filtered version containing just those APIs.

## Features

✨ **Clean & Intuitive Interface**
- Simple, minimalist design
- Full dark mode support (auto-persists preference)
- Responsive layout for all devices

🔄 **Full OpenAPI Support**
- Compatible with OpenAPI 3.0 and Swagger 2.0
- Automatic format detection
- Preserves all original content without modification

🎯 **Smart Module Selection**
- Load OpenAPI from HTTP URL
- Automatic tag extraction and display
- Multi-select modules (tags)
- Select/Deselect all functionality

📥 **Clean Output**
- Download filtered JSON maintaining original structure
- Only removes unrelated API paths
- All schema definitions and constraints preserved

## Installation

### Prerequisites
- Go 1.21 or higher

### Build from Source

```bash
# Clone the repository
git clone https://github.com/CodingOX/openapi-trim.git
cd openapi-trim

# Build the application
go build -o openapi-trim

# Run
./openapi-trim
```

The application will start on `http://localhost:8080`

## Usage

1. **Load Document**: Paste the HTTP URL of your OpenAPI.json file
2. **Parse**: Click "Load & Parse" to fetch and analyze the document
3. **Select Modules**: Choose which API modules (tags) you want to include
4. **Download**: Click "Download Filtered JSON" to get your filtered document

### Example URLs
- `https://api.example.com/openapi.json`
- `https://petstore.swagger.io/v2/swagger.json`
- Any public OpenAPI/Swagger 2.0 document URL

## Project Structure

```
openapi-trim/
├── main.go              # HTTP server and API endpoints
├── openapi/
│   ├── parser.go        # OpenAPI parsing logic (3.0 & 2.0)
│   └── filter.go        # Document filtering by tags
├── web/
│   ├── index.html       # Frontend UI
│   ├── styles.css       # Styling with dark mode
│   └── app.js          # Frontend logic
├── go.mod              # Go dependencies
├── .gitignore          # Git ignore rules
└── README.md           # This file
```

## API Endpoints

### POST /api/parse
Parse an OpenAPI document from a URL.

**Request:**
```json
{
  "url": "https://api.example.com/openapi.json"
}
```

**Response:**
```json
{
  "success": true,
  "message": "OpenAPI parsed successfully",
  "tags": ["users", "products", "orders"],
  "version": "3.0"
}
```

### POST /api/filter
Filter the currently loaded document by selected tags.

**Request:**
```json
{
  "selectedTags": ["users", "products"]
}
```

**Response:**
```json
{
  "success": true,
  "message": "OpenAPI filtered successfully",
  "data": "{...filtered OpenAPI JSON...}"
}
```

## How It Works

1. **Parsing**: Reads OpenAPI document from provided URL, automatically detects version (2.0 or 3.0)
2. **Tag Extraction**: Analyzes all API operations and extracts unique tags
3. **Filtering**: Creates a copy of original document, removes paths not matching selected tags
4. **Output**: Returns filtered JSON with all original constraints and schemas preserved

## Technical Details

### Supported Features
- ✅ OpenAPI 3.0.x
- ✅ Swagger 2.0
- ✅ All HTTP methods (GET, POST, PUT, DELETE, PATCH, etc.)
- ✅ Multi-tag operations (operation with multiple tags)
- ✅ Schema references and definitions
- ✅ Security schemes and authentication
- ✅ Request/response examples

### Dark Mode
Dark mode preference is automatically saved to browser localStorage and restored on next visit.

## Development

### Running in Development

```bash
# Build and run
go run main.go

# Visit http://localhost:8080
```

### Structure of Filtering
The filtering process:
1. Creates deep copy of original OpenAPI document
2. Iterates through all paths and operations
3. Keeps only operations where at least one tag matches selection
4. Preserves all other document elements (info, servers, components, etc.)
5. Returns modified JSON without formatting changes (except indentation)

## Contributing

Contributions are welcome! Feel free to:
- Report bugs
- Suggest features
- Submit pull requests

## License

MIT License - feel free to use and modify as needed.

## Support

For issues and questions, please open an issue on GitHub.

---

**Made with ❤️ for developers who work with large API documentation**
