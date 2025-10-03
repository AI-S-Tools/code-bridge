# Code-Bridge Projektplan

## 1. Projektoversigt

Code-Bridge er et værktøj til at skabe et søgbart RAG (Retrieval-Augmented Generation) indeks over en komplet kodebase. Systemet scanner alle filer rekursivt, ekstraherer funktioner og klasser, og gemmer dem i et JSONL-format for hurtige opslag.

### Hovedformål
- Indeksere hele kodebasen i et struktureret JSONL-format
- Muliggøre RAG-baserede opslag med præcise fil- og linjereferencer
- Tilbyde søgefunktionalitet (generisk og specifik)
- Understøtte annotation af funktioner og klasser uden at ændre kildekode

---

## 2. Systemarkitektur

### 2.1 Komponenter

```
code-bridge/
├── .code-bridge/           # Data og konfiguration
│   ├── codebase.jsonl      # Hovedindeksfil
│   ├── annotations.jsonl   # Brugerannotationer
│   └── config.json         # Konfiguration
├── src/
│   ├── scanner/            # Filscanning og parsing
│   ├── indexer/            # JSONL indeksering
│   ├── rag/                # RAG motor
│   ├── search/             # Søgefunktionalitet
│   └── api/                # API/CLI interface
└── docs/                   # Dokumentation
```

### 2.2 Dataflow

```
Kodebase → Scanner → Parser → Indexer → JSONL → RAG/Search → Resultat
                                          ↓
                                    Annotations
```

---

## 3. Funktionelle Krav

### 3.1 Filscanning og Indeksering

**Input:** Start-folder sti
**Output:** codebase.jsonl

**Funktionalitet:**
- Rekursiv scanning af alle filer fra start-folder
- Filtrering baseret på konfiguration (ignorér node_modules, .git, etc.)
- Parsing af kode for at ekstrahere:
  - Funktioner (navn, parametre, returtype, lokation)
  - Klasser (navn, metoder, attributter, lokation)
  - Imports og dependencies
- Gemme i JSONL-format med metadata

**JSONL Format:**
```jsonl
{"type":"function","name":"calculateTotal","file":"src/utils/math.js","line":15,"end_line":23,"params":["items","tax"],"returns":"number","body":"...","hash":"abc123"}
{"type":"class","name":"UserService","file":"src/services/user.js","line":10,"end_line":150,"methods":[...],"hash":"def456"}
```

### 3.2 RAG (Retrieval-Augmented Generation) System

**Funktionalitet:**
- Vektorisering af kode og dokumentation
- Semantisk søgning baseret på natural language queries
- Returnere relevante kodesegmenter med præcise referencer
- Kontekstuel forståelse af koderelationer

**Opslag Format:**
```python
rag.search("find function that calculates totals")
# Returns:
# {
#   "matches": [
#     {
#       "type": "function",
#       "name": "calculateTotal",
#       "file": "src/utils/math.js",
#       "line": 15,
#       "relevance": 0.95,
#       "context": "..."
#     }
#   ]
# }
```

### 3.3 Søgefunktionalitet

**A. Generisk Søgning:**
- Fritekst søgning i funktions- og klassenavne
- Søgning i kommentarer og docstrings
- Fuzzy matching

**B. Specifik Søgning:**
- Efter funktionsnavn (præcis match)
- Efter input-parametre (type og antal)
- Efter output/returtype
- Efter fil eller mappe
- Efter annotations

**Eksempler:**
```javascript
// Generisk
search.query("user authentication")

// Specifik
search.function({
  name: "authenticate",
  params: ["username", "password"],
  returns: "boolean"
})

// Combined
search.advanced({
  type: "function",
  file: "src/auth/*.js",
  hasAnnotation: "reviewed"
})
```

### 3.4 Annotationssystem

**Funktionalitet:**
- Tilføj metadata til funktioner og klasser uden at ændre kildekode
- Annotations gemmes separat i annotations.jsonl
- Knyttes til kode via hash eller fil+linje reference
- Understøtter tags, noter, status, osv.

**Annotation Format:**
```jsonl
{"target":"calculateTotal@src/utils/math.js:15","tags":["reviewed","critical"],"notes":"Needs optimization","status":"approved","author":"user1","timestamp":"2025-10-03T10:00:00Z"}
```

**API:**
```javascript
annotate.add({
  target: "calculateTotal",
  file: "src/utils/math.js",
  tags: ["reviewed"],
  notes: "Needs optimization"
})

annotate.get("calculateTotal")
// Returns all annotations for function
```

---

## 4. Teknisk Implementering

### 4.1 Scanner Module

**Ansvar:** Rekursiv filtraversering og filtrering

**Teknologier:**
- Node.js `fs.promises` for async file operations
- Glob patterns for fil matching
- `.gitignore` respekt

**Pseudokode:**
```javascript
async function scanDirectory(startPath, config) {
  const files = []
  const queue = [startPath]

  while (queue.length > 0) {
    const dir = queue.shift()
    const entries = await fs.readdir(dir, { withFileTypes: true })

    for (const entry of entries) {
      const fullPath = path.join(dir, entry.name)

      if (shouldIgnore(fullPath, config)) continue

      if (entry.isDirectory()) {
        queue.push(fullPath)
      } else {
        files.push(fullPath)
      }
    }
  }

  return files
}
```

### 4.2 Parser Module

**Ansvar:** Kodeanalyse og ekstraktion

**Teknologier:**
- Tree-sitter (multi-language parsing)
- Babel parser (JavaScript/TypeScript)
- Python AST (Python)
- Fallback: Regex-baseret ekstraktion

**Per-sprog implementering:**
- JavaScript/TypeScript: Babel + TypeScript compiler API
- Python: ast module
- Go: go/parser
- Java: JavaParser
- Generisk: Tree-sitter

**Pseudokode:**
```javascript
async function parseFile(filePath) {
  const content = await fs.readFile(filePath, 'utf-8')
  const ext = path.extname(filePath)
  const parser = getParserForExtension(ext)

  const ast = parser.parse(content)
  const elements = []

  // Ekstraher funktioner
  traverseAST(ast, {
    FunctionDeclaration(node) {
      elements.push({
        type: 'function',
        name: node.name,
        file: filePath,
        line: node.loc.start.line,
        end_line: node.loc.end.line,
        params: extractParams(node),
        returns: extractReturnType(node),
        body: content.slice(node.start, node.end),
        hash: hashCode(content.slice(node.start, node.end))
      })
    },
    ClassDeclaration(node) {
      // Similar for classes
    }
  })

  return elements
}
```

### 4.3 Indexer Module

**Ansvar:** JSONL fil håndtering

**Funktionalitet:**
- Skriv til codebase.jsonl (append-only for performance)
- Incremental updates (kun ændrede filer)
- Deduplication baseret på hash
- Komprimering/arkivering af gamle entries

**Pseudokode:**
```javascript
class Indexer {
  async index(elements) {
    const stream = fs.createWriteStream('.code-bridge/codebase.jsonl', { flags: 'a' })

    for (const element of elements) {
      // Check if already indexed
      if (await this.exists(element.hash)) continue

      stream.write(JSON.stringify(element) + '\n')
    }

    stream.end()
  }

  async exists(hash) {
    // Check existing index for hash
    // Use in-memory bloom filter for fast lookup
  }

  async rebuild() {
    // Re-scan entire codebase
    // Useful when schema changes
  }
}
```

### 4.4 RAG Module

**Ansvar:** Semantisk søgning og retrieval

**Teknologier:**
- Embedding model: sentence-transformers, OpenAI, eller lokal model
- Vector DB: Qdrant, Pinecone, eller in-memory FAISS
- LLM integration: Optional for enhanced results

**Workflow:**
1. Ved indeksering: Generer embeddings for hver kode entry
2. Gem embeddings i vector DB med metadata
3. Ved søgning: Embed query → vector search → return top-k med references

**Pseudokode:**
```javascript
class RAGEngine {
  constructor(embeddingModel, vectorDB) {
    this.embedder = embeddingModel
    this.db = vectorDB
  }

  async addToIndex(element) {
    // Kombiner navn, parametre, og body til embedding
    const text = `${element.name} ${element.params.join(' ')} ${element.body}`
    const embedding = await this.embedder.embed(text)

    await this.db.insert({
      id: element.hash,
      vector: embedding,
      metadata: element
    })
  }

  async search(query, topK = 5) {
    const queryEmbedding = await this.embedder.embed(query)
    const results = await this.db.search(queryEmbedding, topK)

    return results.map(r => ({
      ...r.metadata,
      relevance: r.score
    }))
  }
}
```

### 4.5 Search Module

**Ansvar:** Direkte og kompleks søgning

**Funktionalitet:**
- Query parser for avancerede søgninger
- Indeks for hurtig opslag (in-memory eller SQLite)
- Filtering og ranking

**Pseudokode:**
```javascript
class SearchEngine {
  async search(query) {
    if (typeof query === 'string') {
      return this.genericSearch(query)
    } else {
      return this.advancedSearch(query)
    }
  }

  async genericSearch(text) {
    // Fuzzy search i navn og body
    const lines = await this.readJSONL('.code-bridge/codebase.jsonl')
    return lines.filter(el =>
      el.name.includes(text) ||
      el.body.includes(text)
    )
  }

  async advancedSearch({ name, params, returns, file, annotations }) {
    let results = await this.getAllElements()

    if (name) results = results.filter(el => el.name === name)
    if (params) results = results.filter(el =>
      arraysEqual(el.params, params)
    )
    if (returns) results = results.filter(el => el.returns === returns)
    if (file) results = results.filter(el =>
      minimatch(el.file, file)
    )
    if (annotations) {
      const annotated = await this.getAnnotatedElements(annotations)
      results = results.filter(el => annotated.has(el.hash))
    }

    return results
  }
}
```

### 4.6 Annotation Module

**Ansvar:** Håndtering af annotations uden kodefiler ændring

**Storage:** annotations.jsonl
- Mapping: code element hash → annotations
- Supports multiple annotations per element
- Timestamps og versioning

**Pseudokode:**
```javascript
class AnnotationManager {
  async add({ target, file, line, tags, notes, status }) {
    // Find target element
    const element = await this.findElement(target, file, line)

    const annotation = {
      target: `${element.name}@${element.file}:${element.line}`,
      hash: element.hash,
      tags: tags || [],
      notes: notes || '',
      status: status || 'draft',
      author: getCurrentUser(),
      timestamp: new Date().toISOString()
    }

    await this.appendToFile('.code-bridge/annotations.jsonl', annotation)
    return annotation
  }

  async get(target) {
    const element = await this.findElement(target)
    const lines = await this.readJSONL('.code-bridge/annotations.jsonl')

    return lines.filter(a => a.hash === element.hash)
  }

  async update(annotationId, updates) {
    // Read all, update matching, write back
    // Or use append-only with latest-wins logic
  }
}
```

---

## 5. API & CLI Design

### 5.1 CLI Commands

```bash
# Initialize
code-bridge init [path]

# Index codebase
code-bridge index [--incremental]

# Search
code-bridge search "query"
code-bridge search --function calculateTotal --params items,tax
code-bridge search --file "src/utils/*" --type function

# RAG query
code-bridge rag "find functions that handle user authentication"

# Annotations
code-bridge annotate add --target calculateTotal --tags reviewed,critical
code-bridge annotate list calculateTotal
code-bridge annotate search --tag reviewed

# Maintenance
code-bridge rebuild
code-bridge stats
```

### 5.2 Programmatic API

```javascript
const CodeBridge = require('code-bridge')

// Initialize
const cb = new CodeBridge('/path/to/codebase')
await cb.init()

// Index
await cb.index({ incremental: true })

// Search
const results = await cb.search.function({
  name: 'calculateTotal',
  params: ['items', 'tax']
})

// RAG
const ragResults = await cb.rag.query(
  'find functions that calculate totals'
)

// Annotations
await cb.annotate.add({
  target: 'calculateTotal',
  tags: ['reviewed'],
  notes: 'Needs optimization'
})

const annotations = await cb.annotate.get('calculateTotal')
```

### 5.3 REST API (Optional)

```
POST   /api/index              # Trigger indexing
GET    /api/search?q=...       # Search
POST   /api/rag                # RAG query
GET    /api/functions/:name    # Get function details
POST   /api/annotations        # Add annotation
GET    /api/annotations/:target # Get annotations
```

---

## 6. Data Formats

### 6.1 codebase.jsonl Schema

```typescript
interface CodeElement {
  type: 'function' | 'class' | 'interface' | 'type' | 'variable'
  name: string
  file: string              // Relative path
  line: number             // Start line
  end_line: number         // End line
  hash: string             // Content hash for dedup

  // Function specific
  params?: Array<{
    name: string
    type?: string
    default?: any
  }>
  returns?: string
  async?: boolean
  generator?: boolean

  // Class specific
  methods?: string[]
  extends?: string
  implements?: string[]

  // Common
  body: string             // Full source code
  docstring?: string       // Extracted documentation
  imports?: string[]       // Dependencies
  exports?: boolean

  // Metadata
  language: string
  indexed_at: string       // ISO timestamp
}
```

### 6.2 annotations.jsonl Schema

```typescript
interface Annotation {
  id: string                    // UUID
  target: string               // "functionName@file:line"
  hash: string                 // Code element hash

  tags: string[]               // ["reviewed", "critical", "deprecated"]
  notes: string                // Free text
  status: 'draft' | 'approved' | 'rejected'
  priority?: 'low' | 'medium' | 'high'

  author: string
  created_at: string           // ISO timestamp
  updated_at?: string

  // Optional references
  related?: string[]           // Other element hashes
  issues?: string[]            // Issue tracker IDs
}
```

### 6.3 config.json Schema

```typescript
interface Config {
  root: string                 // Project root path
  include: string[]            // Glob patterns to include
  exclude: string[]            // Glob patterns to exclude

  languages: string[]          // Enabled languages

  indexing: {
    incremental: boolean       // Only re-index changed files
    watch: boolean            // Watch for file changes
    parallel: boolean         // Parallel processing
  }

  rag: {
    enabled: boolean
    model: string             // Embedding model
    provider: 'local' | 'openai' | 'custom'
    api_key?: string
    vector_db: {
      type: 'qdrant' | 'faiss' | 'pinecone'
      config: Record<string, any>
    }
  }

  search: {
    fuzzy: boolean
    max_results: number
  }
}
```

---

## 7. Implementeringsfaser

### **Fase 1: Core Infrastructure (Uge 1-2)**
- [ ] Projekt setup og arkitektur
- [ ] Scanner implementering (rekursiv fil traversering)
- [ ] Basic parser (JavaScript/TypeScript support)
- [ ] JSONL indexer (write/read operations)
- [ ] CLI grundlæggende struktur

### **Fase 2: Search & Retrieval (Uge 3)**
- [ ] Generisk søgning i indeks
- [ ] Specifik søgning (navn, params, returns)
- [ ] Fil og path filtering
- [ ] Performance optimering (in-memory indeks)

### **Fase 3: RAG Integration (Uge 4)**
- [ ] Embedding model integration
- [ ] Vector database setup
- [ ] Semantisk søgning
- [ ] Relevance ranking

### **Fase 4: Annotation System (Uge 5)**
- [ ] Annotation storage (JSONL)
- [ ] Add/update/delete annotations
- [ ] Annotation search
- [ ] UI for annotation management

### **Fase 5: Multi-language Support (Uge 6)**
- [ ] Python parser
- [ ] Go parser
- [ ] Java parser
- [ ] Generisk fallback parser

### **Fase 6: Advanced Features (Uge 7-8)**
- [ ] Incremental indexing
- [ ] File watching
- [ ] Code change detection
- [ ] Relation mapping (function calls, imports)
- [ ] Visualization/reporting

### **Fase 7: API & Integration (Uge 9)**
- [ ] REST API
- [ ] WebSocket support (real-time)
- [ ] IDE plugins (VSCode)
- [ ] CI/CD integration

### **Fase 8: Testing & Documentation (Uge 10)**
- [ ] Unit tests (>80% coverage)
- [ ] Integration tests
- [ ] Performance benchmarks
- [ ] Dokumentation
- [ ] Eksempler og tutorials

---

## 8. Teknologi Stack

### Core
- **Runtime:** Node.js 18+ eller Deno
- **Language:** TypeScript
- **CLI:** Commander.js eller Yargs

### Parsing
- **Multi-lang:** Tree-sitter
- **JavaScript/TS:** Babel + TypeScript Compiler API
- **Python:** Python ast module (via child process)
- **Go:** go/parser (via child process)

### Storage
- **Index:** JSONL filer
- **Cache:** SQLite eller LevelDB
- **Vector DB:** Qdrant (self-hosted) eller FAISS (local)

### RAG
- **Embeddings:** sentence-transformers (local) eller OpenAI
- **Inference:** Optional LLM integration

### Development
- **Testing:** Vitest eller Jest
- **Bundling:** esbuild
- **Linting:** ESLint + Prettier

---

## 9. Performance Targets

- **Indexing:** >1000 filer/sekund (medium-sized codebase)
- **Search:** <50ms response tid
- **RAG Query:** <200ms response tid
- **Memory:** <500MB for 100k code elements
- **Storage:** ~1KB per code element

---

## 10. Security & Privacy

- Lokal-first: Alle data gemmes lokalt
- API keys krypteret i config
- Option for cloud sync (encrypted)
- No telemetry uden explicit opt-in
- .gitignore respekt for sensitive filer

---

## 11. Testing Strategy

### Unit Tests
- Parser tests per sprog
- Indexer read/write operations
- Search query parsing
- Annotation CRUD

### Integration Tests
- End-to-end indexing pipeline
- Search accuracy
- RAG relevance
- CLI commands

### Performance Tests
- Large codebase indexing (>100k files)
- Concurrent search operations
- Memory profiling
- Storage efficiency

---

## 12. Fremtidige Udvidelser

- **Code Intelligence:**
  - Call graph generation
  - Dependency analysis
  - Dead code detection

- **Collaboration:**
  - Team annotations
  - Code review integration
  - Shared knowledge base

- **AI Features:**
  - Code generation from RAG context
  - Automatic documentation
  - Bug prediction

- **Integrations:**
  - GitHub/GitLab
  - Jira/Linear
  - Slack/Discord notifications

---

## 13. Success Metrics

- Indexing hastighed vs. codebase størrelse
- Search relevance (precision/recall)
- RAG accuracy (bruger feedback)
- Adoption rate (downloads, stars)
- Performance benchmarks vs. alternativer (Sourcegraph, OpenGrok)

---

## 14. Risks & Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| Parser support for alle sprog | High | Start med top 5 sprog, fallback til regex |
| RAG model størrelse/performance | Medium | Offer både local og cloud options |
| Storage vækst for store codebases | Medium | Compression, archiving, selective indexing |
| Search relevance kvalitet | High | Combine text + semantic search, user feedback |
| Breaking changes i AST parsers | Low | Version lock dependencies, fallback strategies |

---

## 15. Konklusion

Code-Bridge er et ambitiøst men realiserbart projekt, der kan revolutionere hvordan udviklere navigerer og forstår store codebases. Ved at kombinere traditionel indeksering med moderne RAG-teknologi, tilbyder det både præcision og semantisk forståelse.

**Next Steps:**
1. Godkendelse af projektplan
2. Setup development environment
3. Implementer Fase 1 (Core Infrastructure)
4. Første prototype klar til testing

**Estimeret tid til MVP:** 6-8 uger
**Estimeret tid til production-ready:** 10-12 uger
