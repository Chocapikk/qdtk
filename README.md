# qdtk - Qdrant ToolKit

<p align="center">
  <b>A standalone solution to navigate and backup data from Qdrant vector databases</b>
</p>

<p align="center">
  <i>Inspired by <a href="https://github.com/LeakIX/estk">estk</a> from LeakIX</i>
</p>

---

## Features

- **List** - Browse collections with point counts
- **Stats** - Get detailed database statistics
- **Dump** - Full backup of collections to JSON (supports all collections at once)
- **Search** - Text search across payload fields
- **TLS Support** - Works with HTTPS endpoints (skips certificate verification)
- **Progress Bar** - Visual feedback during large dumps
- **Resumable** - Scroll-based pagination for reliable large exports

## Installation

### From source

```bash
git clone https://github.com/chocapikk/qdtk.git
cd qdtk
go build -o qdtk .
```

### Pre-built binaries

Coming soon on the releases page.

## Usage

### List collections

```bash
# Simple list
qdtk --url="https://qdrant.example.com" list

# Verbose with details
qdtk --url="https://qdrant.example.com" list -v
```

### Get statistics

```bash
qdtk --url="https://qdrant.example.com" stats
```

Output:
```
▸ Database Info
──────────────────────────────────────────────────
  Version: Qdrant 1.7.4
  Collections: 8

▸ Collection Details
──────────────────────────────────────────────────
Collection          Points              Vectors             Status
────────────────────────────────────────────────────────────────────
users               10726               10726               green
orders              5429                5429                green

──────────────────────────────────────────────────
  Total Points: 16155
  Total Vectors: 16155
```

### Dump data

```bash
# Dump single collection
qdtk --url="https://qdrant.example.com" dump -c users -o users.json

# Dump ALL collections
qdtk --url="https://qdrant.example.com" dump -c "*" -o full_backup.json

# Dump with limit
qdtk --url="https://qdrant.example.com" dump -c users -l 1000 -o sample.json

# Dump payload only (no metadata)
qdtk --url="https://qdrant.example.com" dump -c users -p -o users_payload.json

# Include vectors in dump
qdtk --url="https://qdrant.example.com" dump -c users -v -o users_with_vectors.json

# Pipe to stdout
qdtk --url="https://qdrant.example.com" dump -c users | jq '.payload'
```

### Search

```bash
# Search text in all fields
qdtk --url="https://qdrant.example.com" search -c users -q "admin"

# Search in specific field
qdtk --url="https://qdrant.example.com" search -c users -q "password" -f text

# Raw JSON output
qdtk --url="https://qdrant.example.com" search -c users -q "secret" -r
```

## Output Format

Each dumped point contains:

```json
{
  "_collection": "users",
  "_id": "550e8400-e29b-41d4-a716-446655440000",
  "payload": {
    "email": "user@example.com",
    "name": "John Doe"
  }
}
```

With `-p` flag (payload only):

```json
{
  "email": "user@example.com",
  "name": "John Doe"
}
```

## Command Reference

| Command | Description |
|---------|-------------|
| `list` | List all collections |
| `stats` | Show database statistics |
| `dump` | Dump collection data to JSON |
| `search` | Search within collection payloads |

### Global Flags

| Flag | Description |
|------|-------------|
| `--url` | Qdrant URL (required) |
| `-d, --debug` | Enable debug mode |

### Dump Flags

| Flag | Description | Default |
|------|-------------|---------|
| `-c, --collection` | Collection name (`*` for all) | required |
| `-o, --output` | Output file | stdout |
| `-l, --limit` | Max documents (0 = unlimited) | 0 |
| `-b, --batch-size` | Scroll batch size | 100 |
| `-v, --with-vectors` | Include vectors | false |
| `-p, --payload-only` | Output payload only | false |
| `-q, --quiet` | Minimal output | false |

### Search Flags

| Flag | Description | Default |
|------|-------------|---------|
| `-c, --collection` | Collection name | required |
| `-q, --query` | Search text | "" |
| `-f, --field` | Specific field to search | all |
| `-l, --limit` | Max results | 10 |
| `-r, --raw` | Raw JSON output | false |

## Use Cases

### Security Research

Identify exposed Qdrant instances and assess data exposure:

```bash
# Quick assessment
qdtk --url="https://target:6333" stats

# Full backup for analysis
qdtk --url="https://target:6333" dump -c "*" -o backup.json

# Search for sensitive data
qdtk --url="https://target:6333" search -c users -q "password"
qdtk --url="https://target:6333" search -c users -q "@gmail.com"
```

### Data Migration

Export data from one Qdrant instance:

```bash
qdtk --url="https://source:6333" dump -c my_collection -o export.json
```

### Backup

Regular backups of your Qdrant data:

```bash
qdtk --url="https://qdrant:6333" dump -c "*" -o "backup_$(date +%Y%m%d).json"
```

## Credits

- Inspired by [estk](https://github.com/LeakIX/estk) from [LeakIX](https://leakix.net)
- Built for the security research community

## License

MIT License - See [LICENSE](LICENSE) for details.

## Disclaimer

This tool is intended for authorized security testing and legitimate data management purposes only. Always ensure you have proper authorization before accessing any Qdrant instance.
