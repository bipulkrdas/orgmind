import type { Metadata, Viewport } from 'next';
import './globals.css';

export const metadata: Metadata = {
  title: 'OrgMind - Document Processing Platform',
  description: 'Enterprise document processing platform with AI-powered knowledge graph generation. Transform your documents into actionable insights.',
  keywords: ['document processing', 'knowledge graph', 'AI', 'enterprise', 'document management'],
  authors: [{ name: 'OrgMind' }],
  icons: {
    icon: './icon.png',
  },
};

export const viewport: Viewport = {
  width: 'device-width',
  initialScale: 1,
  themeColor: '#4F46E5',
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en" className="h-full">
      <body className="h-full antialiased">{children}</body>
    </html>
  );
}
