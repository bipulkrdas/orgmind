'use client';

import { useLexicalComposerContext } from '@lexical/react/LexicalComposerContext';
import { useEffect } from 'react';
import { $getRoot } from 'lexical';

const AUTO_SAVE_KEY = 'orgmind-editor-autosave';
const AUTO_SAVE_INTERVAL = 2000; // 2 seconds

export default function AutoSavePlugin() {
  const [editor] = useLexicalComposerContext();

  useEffect(() => {
    // Load saved content on mount
    const savedContent = localStorage.getItem(AUTO_SAVE_KEY);
    if (savedContent) {
      try {
        const editorState = editor.parseEditorState(savedContent);
        editor.setEditorState(editorState);
      } catch (error) {
        console.error('Failed to load saved content:', error);
      }
    }
  }, [editor]);

  useEffect(() => {
    let timeoutId: NodeJS.Timeout;

    const saveContent = () => {
      editor.getEditorState().read(() => {
        const root = $getRoot();
        const textContent = root.getTextContent();
        
        // Only save if there's content
        if (textContent.trim()) {
          const editorStateJSON = JSON.stringify(editor.getEditorState().toJSON());
          localStorage.setItem(AUTO_SAVE_KEY, editorStateJSON);
        }
      });
    };

    // Set up auto-save on editor updates
    const removeUpdateListener = editor.registerUpdateListener(() => {
      clearTimeout(timeoutId);
      timeoutId = setTimeout(saveContent, AUTO_SAVE_INTERVAL);
    });

    return () => {
      clearTimeout(timeoutId);
      removeUpdateListener();
    };
  }, [editor]);

  return null;
}
