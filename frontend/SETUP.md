# Frontend Setup Instructions

## Initial Setup

The frontend project structure has been created with all necessary configuration files.

### Fix npm Cache Issue (if needed)

If you encounter npm cache permission errors, run:

```bash
sudo chown -R $(whoami) ~/.npm
```

### Install Dependencies

After fixing the cache issue, install dependencies:

```bash
cd frontend
npm install
```

### Start Development Server

```bash
npm run dev
```

The application will be available at [http://localhost:3000](http://localhost:3000)

## Project Structure Created

✅ Next.js 14+ with TypeScript and Tailwind CSS
✅ App Router with route groups:
  - `(public)` - Unauthenticated routes
  - `(auth)` - Authenticated routes
✅ Component directories for:
  - Lexical editor
  - File upload
  - Graph visualization
  - Authentication forms
✅ Library directories for:
  - API client functions
  - JWT token management
  - TypeScript types
✅ Environment configuration files
✅ Middleware for route protection (placeholder)

## Dependencies Configured

The following dependencies are configured in package.json:

- **Core**: next, react, react-dom
- **Editor**: lexical, @lexical/react, @lexical/rich-text, @lexical/list, @lexical/link, @lexical/utils
- **Graph**: sigma, graphology
- **State Management**: @tanstack/react-query
- **Styling**: tailwindcss
- **TypeScript**: typescript, @types/node, @types/react, @types/react-dom

## Next Steps

After running `npm install`, the frontend will be ready for implementing:
- Task 14: TypeScript types and API utilities
- Task 15: Authentication components
- Task 16: Public route pages
- Task 17: Route protection middleware
- Task 18: Lexical editor component
- Task 19: File upload component
- Task 20: Authenticated home page
- Task 21: Graph visualization component
