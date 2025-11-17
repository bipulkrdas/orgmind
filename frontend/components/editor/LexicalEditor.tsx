'use client';

import { useEffect, useState } from 'react';
import { LexicalComposer } from '@lexical/react/LexicalComposer';
import { useLexicalComposerContext } from '@lexical/react/LexicalComposerContext';
import { RichTextPlugin } from '@lexical/react/LexicalRichTextPlugin';
import { ContentEditable } from '@lexical/react/LexicalContentEditable';
import { HistoryPlugin } from '@lexical/react/LexicalHistoryPlugin';
import { AutoFocusPlugin } from '@lexical/react/LexicalAutoFocusPlugin';
import { HeadingNode, QuoteNode } from '@lexical/rich-text';
import { ListItemNode, ListNode } from '@lexical/list';
import { LinkNode } from '@lexical/link';
import { ListPlugin } from '@lexical/react/LexicalListPlugin';
import { OnChangePlugin } from '@lexical/react/LexicalOnChangePlugin';
import LexicalErrorBoundary from '@lexical/react/LexicalErrorBoundary';
import { EditorState, $getRoot, $createParagraphNode, $createTextNode } from 'lexical';

import ToolbarPlugin from '@/components/editor/ToolbarPlugin';
import AutoSavePlugin from '@/components/editor/AutoSavePlugin';

interface LexicalEditorProps {
  onSubmit: (plainText: string, lexicalState: string) => Promise<void>;
  initialContent?: string;
  initialLexicalState?: string;
}

const theme = {
  paragraph: 'mb-2',
  heading: {
    h1: 'text-3xl font-bold mb-4',
    h2: 'text-2xl font-bold mb-3',
    h3: 'text-xl font-bold mb-2',
  },
  list: {
    ul: 'list-disc list-inside mb-2',
    ol: 'list-decimal list-inside mb-2',
    listitem: 'ml-4',
  },
  text: {
    bold: 'font-bold',
    italic: 'italic',
    underline: 'underline',
  },
};

// Component to capture editor reference
function EditorRefPlugin({ setEditorRef }: { setEditorRef: (editor: any) => void }) {
  const [editor] = useLexicalComposerContext();
  
  useEffect(() => {
    setEditorRef(editor);
  }, [editor, setEditorRef]);
  
  return null;
}

export default function LexicalEditor({ onSubmit, initialContent, initialLexicalState }: LexicalEditorProps) {
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState(false);
  const [editorState, setEditorState] = useState<EditorState | null>(null);
  const [editor, setEditor] = useState<any>(null);

  const initialConfig = {
    namespace: 'OrgMindEditor',
    theme,
    onError: (error: Error) => {
      console.error('Lexical error:', error);
    },
    nodes: [
      HeadingNode,
      QuoteNode,
      ListNode,
      ListItemNode,
      LinkNode,
    ],
  };

  // Load initial content into editor when editor is ready
  useEffect(() => {
    if (!editor) return;

    // Priority 1: Load from Lexical state (preserves formatting)
    if (initialLexicalState) {
      try {
        const parsedState = editor.parseEditorState(initialLexicalState);
        editor.setEditorState(parsedState);
        return;
      } catch (err) {
        console.error('Failed to parse Lexical state:', err);
        // Fall through to plain text loading
      }
    }

    // Priority 2: Load from plain text (fallback)
    if (initialContent) {
      editor.update(() => {
        const root = $getRoot();
        root.clear();
        const paragraph = $createParagraphNode();
        const textNode = $createTextNode(initialContent);
        paragraph.append(textNode);
        root.append(paragraph);
      });
    }
  }, [editor, initialContent, initialLexicalState]);

  const handleEditorChange = (editorState: EditorState) => {
    setEditorState(editorState);
  };

  const handleSubmit = async () => {
    if (!editorState || !editor) {
      setError('No content to submit');
      return;
    }

    setIsSubmitting(true);
    setError(null);
    setSuccess(false);

    try {
      // Extract plain text content for Zep processing
      let plainText = '';
      editorState.read(() => {
        const root = editorState._nodeMap.get('root');
        if (root) {
          plainText = root.getTextContent();
        }
      });

      if (!plainText.trim()) {
        setError('Please enter some content before submitting');
        setIsSubmitting(false);
        return;
      }

      // Serialize Lexical state to JSON (preserves formatting)
      const lexicalStateJSON = JSON.stringify(editorState.toJSON());

      // Submit both versions to backend
      await onSubmit(plainText, lexicalStateJSON);
      setSuccess(true);
      
      // Clear the editor and localStorage after successful submission
      editor.update(() => {
        const root = $getRoot();
        root.clear();
      });
      localStorage.removeItem('orgmind-editor-autosave');
      
      // Clear success message after 3 seconds
      setTimeout(() => setSuccess(false), 3000);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to submit content');
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className="w-full max-w-4xl mx-auto">
      <LexicalComposer initialConfig={initialConfig}>
        <div className="border border-gray-300 rounded-lg shadow-sm bg-white">
          <ToolbarPlugin />
          <div className="relative">
            <RichTextPlugin
              contentEditable={
                <ContentEditable className="min-h-[400px] p-4 outline-none" />
              }
              placeholder={
                <div className="absolute top-4 left-4 text-gray-400 pointer-events-none">
                  Start typing your document...
                </div>
              }
              ErrorBoundary={LexicalErrorBoundary}
            />
          </div>
          <HistoryPlugin />
          <AutoFocusPlugin />
          <ListPlugin />
          <OnChangePlugin onChange={handleEditorChange} />
          <AutoSavePlugin />
          <EditorRefPlugin setEditorRef={setEditor} />
        </div>

        <div className="mt-4 flex items-center justify-between">
          <div className="flex-1">
            {error && (
              <div className="text-red-600 text-sm">{error}</div>
            )}
            {success && (
              <div className="text-green-600 text-sm">Document submitted successfully!</div>
            )}
          </div>
          <button
            onClick={handleSubmit}
            disabled={isSubmitting}
            className={`px-6 py-2 rounded-md font-medium transition-colors ${
              isSubmitting
                ? 'bg-gray-400 cursor-not-allowed'
                : 'bg-blue-600 hover:bg-blue-700 text-white'
            }`}
          >
            {isSubmitting ? 'Submitting...' : 'Submit Document'}
          </button>
        </div>
      </LexicalComposer>
    </div>
  );
}
