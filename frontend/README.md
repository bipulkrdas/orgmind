# OrgMind Frontend

Enterprise document processing platform frontend built with Next.js 14+.

## Tech Stack

- **Framework**: Next.js 14+ (App Router)
- **Language**: TypeScript
- **Styling**: Tailwind CSS
- **Rich Text Editor**: Lexical
- **Graph Visualization**: sigma.js
- **State Management**: React Query (@tanstack/react-query)

## Project Structure

```
frontend/
├── app/
│   ├── (public)/              # Unauthenticated routes
│   │   ├── page.tsx           # Landing page
│   │   ├── signup/            # Sign up page
│   │   ├── signin/            # Sign in page
│   │   ├── reset-password/    # Password reset
│   │   └── auth/callback/     # OAuth callback
│   ├── (auth)/                # Authenticated routes
│   │   ├── layout.tsx         # Auth layout with navigation
│   │   ├── home/              # Home page with editor & upload
│   │   └── graphs/[graphId]/  # Graph visualization
│   ├── layout.tsx             # Root layout
│   └── globals.css            # Global styles
├── components/
│   ├── editor/                # Lexical editor components
│   ├── upload/                # File upload components
│   ├── graph/                 # Graph visualization components
│   └── auth/                  # Authentication components
├── lib/
│   ├── api/                   # API client functions
│   ├── auth/                  # JWT token management
│   └── types/                 # TypeScript type definitions
└── middleware.ts              # Route protection middleware
```

## Getting Started

1. Install dependencies:
```bash
npm install
```

2. Copy environment variables:
```bash
cp .env.local.example .env.local
```

3. Update `.env.local` with your configuration:
```
NEXT_PUBLIC_API_URL=http://localhost:8080
NEXT_PUBLIC_OAUTH_REDIRECT_URL=http://localhost:3000/auth/callback
```

4. Run the development server:
```bash
npm run dev
```

5. Open [http://localhost:3000](http://localhost:3000) in your browser.

## Environment Variables

- `NEXT_PUBLIC_API_URL`: Backend API URL
- `NEXT_PUBLIC_OAUTH_REDIRECT_URL`: OAuth callback URL

## Development

- `npm run dev` - Start development server
- `npm run build` - Build for production
- `npm run start` - Start production server
- `npm run lint` - Run ESLint

## Route Groups

The app uses Next.js route groups for organization:

- `(public)` - Unauthenticated routes (landing, signup, signin)
- `(auth)` - Authenticated routes (home, graphs)

Route groups don't affect the URL structure but help organize code.
