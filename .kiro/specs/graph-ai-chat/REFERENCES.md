# AI Chat Implementation References

This document contains all external references needed for implementing the AI-powered chat feature with Google Gemini File Search.

## Google Gemini File Search

### Official Documentation

#### Primary Resources
- **File Search Overview**  
  https://ai.google.dev/gemini-api/docs/file-search  
  Complete guide to File Search capabilities and use cases

- **File Search Stores Guide**  
  https://ai.google.dev/gemini-api/docs/file-search#file-search-stores  
  Detailed documentation on creating and managing File Search stores

- **REST API Reference**  
  https://ai.google.dev/gemini-api/docs/file-search#rest  
  REST API endpoints and request/response formats

### Go SDK (go-genai)

#### Repository and Code
- **GitHub Repository**  
  https://github.com/googleapis/go-genai  
  Official Go client library for Google Gemini API

- **Complete Example Implementation**  
  https://github.com/googleapis/go-genai/blob/main/examples/filesearchstores/create_upload_and_call_file_search.go  
  Full working example showing:
  - File Search store creation
  - Document upload
  - Query with File Search
  - Response handling

- **SDK Implementation Source**  
  https://github.com/googleapis/go-genai/blob/main/filesearchstores.go  
  Internal implementation details and method signatures

#### Key SDK Methods

```go
// Client creation
client, err := genai.NewClient(ctx, &genai.ClientConfig{
    APIKey: apiKey,
    Backend: genai.BackendGoogleAI,
})

// File Search store operations
store, err := client.CreateFileSearchStore(ctx, &genai.CreateFileSearchStoreRequest{
    DisplayName: "store-name",
})

// Document upload
file, err := client.UploadFile(ctx, &genai.UploadFileRequest{
    File: reader,
    MIMEType: "application/pdf",
})

// Query with File Search
response := client.GenerateContent(ctx, &genai.GenerateContentRequest{
    Model: "gemini-1.5-flash",
    Contents: []*genai.Content{{
        Parts: []*genai.Part{{Text: query}},
    }},
    Tools: []*genai.Tool{{
        FileSearchTool: &genai.FileSearchTool{
            FileSearchStoreIDs: []string{storeID},
        },
    }},
})
```

### Important Notes

1. **Latest SDK**: The go-genai SDK is actively developed. Always check for updates.
2. **File Search is New**: This is a relatively recent feature. Documentation may evolve.
3. **Authentication**: SDK handles OAuth2 and API key authentication automatically.
4. **Streaming**: SDK supports streaming responses for real-time chat.
5. **Error Handling**: Built-in retry logic with exponential backoff.

## Server-Sent Events (SSE)

### Specifications
- **MDN Web Docs**  
  https://developer.mozilla.org/en-US/docs/Web/API/Server-sent_events  
  Complete guide to SSE API and usage

- **HTML5 Specification**  
  https://html.spec.whatwg.org/multipage/server-sent-events.html  
  Official W3C specification

### Go Implementation with Gin

- **Gin SSE Example**  
  https://github.com/gin-gonic/examples/tree/master/server-sent-event  
  Official example from Gin framework

#### Basic SSE Pattern in Gin

```go
func StreamHandler(c *gin.Context) {
    c.Header("Content-Type", "text/event-stream")
    c.Header("Cache-Control", "no-cache")
    c.Header("Connection", "keep-alive")
    
    // Create channel for messages
    messageChan := make(chan string)
    
    // Stream messages
    c.Stream(func(w io.Writer) bool {
        if msg, ok := <-messageChan; ok {
            c.SSEvent("message", msg)
            return true
        }
        return false
    })
}
```

### Frontend SSE Client

```typescript
const eventSource = new EventSource('/api/stream');

eventSource.onmessage = (event) => {
    console.log('Received:', event.data);
};

eventSource.onerror = (error) => {
    console.error('SSE error:', error);
    eventSource.close();
};

// Cleanup
eventSource.close();
```

## Zustand State Management

### Documentation
- **Official Documentation**  
  https://zustand-demo.pmnd.rs/  
  Interactive documentation with examples

- **GitHub Repository**  
  https://github.com/pmndrs/zustand  
  Source code and additional examples

### Basic Usage Pattern

```typescript
import { create } from 'zustand';

interface ChatStore {
    messages: Message[];
    addMessage: (message: Message) => void;
}

const useChatStore = create<ChatStore>((set) => ({
    messages: [],
    addMessage: (message) => 
        set((state) => ({ 
            messages: [...state.messages, message] 
        })),
}));

// In component
const messages = useChatStore((state) => state.messages);
const addMessage = useChatStore((state) => state.addMessage);
```

### Best Practices
1. Use selectors to prevent unnecessary re-renders
2. Keep stores focused and modular
3. Avoid storing derived state
4. Use middleware for persistence if needed

## PostgreSQL and sqlx

### Documentation
- **sqlx Documentation**  
  https://jmoiron.github.io/sqlx/  
  Go database library with struct scanning

- **PostgreSQL JSON Types**  
  https://www.postgresql.org/docs/current/datatype-json.html  
  For storing chat metadata if needed

## Additional Resources

### React and Next.js
- **Next.js App Router**  
  https://nextjs.org/docs/app  
  Latest routing system documentation

- **React Hooks**  
  https://react.dev/reference/react  
  Official React hooks reference

### Tailwind CSS
- **Tailwind Documentation**  
  https://tailwindcss.com/docs  
  For styling chat components

### TypeScript
- **TypeScript Handbook**  
  https://www.typescriptlang.org/docs/handbook/intro.html  
  For type definitions and interfaces

## Implementation Checklist

When implementing, refer to:

- [ ] Google Gemini File Search docs for API details
- [ ] go-genai example code for implementation patterns
- [ ] SSE specifications for streaming setup
- [ ] Zustand docs for state management patterns
- [ ] Existing OrgMind code for architectural consistency

## Quick Links Summary

| Resource | URL |
|----------|-----|
| Gemini File Search Overview | https://ai.google.dev/gemini-api/docs/file-search |
| File Search Stores | https://ai.google.dev/gemini-api/docs/file-search#file-search-stores |
| go-genai Repository | https://github.com/googleapis/go-genai |
| go-genai Example | https://github.com/googleapis/go-genai/blob/main/examples/filesearchstores/create_upload_and_call_file_search.go |
| SSE MDN Docs | https://developer.mozilla.org/en-US/docs/Web/API/Server-sent_events |
| Zustand Docs | https://zustand-demo.pmnd.rs/ |

## Notes for Implementation

1. **Start with the Example**: The go-genai example code is the best starting point
2. **Test Incrementally**: Test File Search store creation before document upload
3. **Handle Errors**: Gemini API may have rate limits and quotas
4. **Monitor Costs**: File Search operations have associated costs
5. **Follow Patterns**: Use existing OrgMind patterns for consistency

## Support and Community

- **Google AI Studio**: https://aistudio.google.com/ - Test Gemini API interactively
- **Stack Overflow**: Tag questions with `google-gemini` and `go`
- **GitHub Issues**: Report SDK issues on go-genai repository
