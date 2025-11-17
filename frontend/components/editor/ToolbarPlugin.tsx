'use client';

import { useLexicalComposerContext } from '@lexical/react/LexicalComposerContext';
import { useCallback, useEffect, useState } from 'react';
import {
  $getSelection,
  $isRangeSelection,
  FORMAT_TEXT_COMMAND,
  SELECTION_CHANGE_COMMAND,
  FORMAT_ELEMENT_COMMAND,
} from 'lexical';
import { $setBlocksType } from '@lexical/selection';
import { $createHeadingNode, $createQuoteNode, HeadingTagType } from '@lexical/rich-text';
import { INSERT_UNORDERED_LIST_COMMAND, INSERT_ORDERED_LIST_COMMAND } from '@lexical/list';
import { mergeRegister } from '@lexical/utils';

export default function ToolbarPlugin() {
  const [editor] = useLexicalComposerContext();
  const [isBold, setIsBold] = useState(false);
  const [isItalic, setIsItalic] = useState(false);
  const [blockType, setBlockType] = useState('paragraph');

  const updateToolbar = useCallback(() => {
    const selection = $getSelection();
    if ($isRangeSelection(selection)) {
      setIsBold(selection.hasFormat('bold'));
      setIsItalic(selection.hasFormat('italic'));
    }
  }, []);

  useEffect(() => {
    return mergeRegister(
      editor.registerUpdateListener(({ editorState }) => {
        editorState.read(() => {
          updateToolbar();
        });
      }),
      editor.registerCommand(
        SELECTION_CHANGE_COMMAND,
        () => {
          updateToolbar();
          return false;
        },
        1
      )
    );
  }, [editor, updateToolbar]);

  const formatHeading = (headingSize: HeadingTagType) => {
    editor.update(() => {
      const selection = $getSelection();
      if ($isRangeSelection(selection)) {
        $setBlocksType(selection, () => $createHeadingNode(headingSize));
      }
    });
  };

  const formatParagraph = () => {
    editor.update(() => {
      const selection = $getSelection();
      if ($isRangeSelection(selection)) {
        $setBlocksType(selection, () => $createQuoteNode());
      }
    });
  };

  const formatBulletList = () => {
    editor.dispatchCommand(INSERT_UNORDERED_LIST_COMMAND, undefined);
  };

  const formatNumberedList = () => {
    editor.dispatchCommand(INSERT_ORDERED_LIST_COMMAND, undefined);
  };

  return (
    <div className="flex items-center gap-1 p-2 border-b border-gray-300 flex-wrap">
      {/* Heading Buttons */}
      <select
        className="px-2 py-1 border border-gray-300 rounded text-sm"
        onChange={(e) => {
          const value = e.target.value;
          if (value === 'h1' || value === 'h2' || value === 'h3') {
            formatHeading(value);
          }
        }}
        value={blockType}
      >
        <option value="paragraph">Normal</option>
        <option value="h1">Heading 1</option>
        <option value="h2">Heading 2</option>
        <option value="h3">Heading 3</option>
      </select>

      <div className="w-px h-6 bg-gray-300 mx-1" />

      {/* Text Format Buttons */}
      <button
        onClick={() => {
          editor.dispatchCommand(FORMAT_TEXT_COMMAND, 'bold');
        }}
        className={`px-3 py-1 rounded text-sm font-bold ${
          isBold ? 'bg-gray-300' : 'hover:bg-gray-100'
        }`}
        aria-label="Format Bold"
      >
        B
      </button>

      <button
        onClick={() => {
          editor.dispatchCommand(FORMAT_TEXT_COMMAND, 'italic');
        }}
        className={`px-3 py-1 rounded text-sm italic ${
          isItalic ? 'bg-gray-300' : 'hover:bg-gray-100'
        }`}
        aria-label="Format Italic"
      >
        I
      </button>

      <button
        onClick={() => {
          editor.dispatchCommand(FORMAT_TEXT_COMMAND, 'underline');
        }}
        className="px-3 py-1 rounded text-sm underline hover:bg-gray-100"
        aria-label="Format Underline"
      >
        U
      </button>

      <div className="w-px h-6 bg-gray-300 mx-1" />

      {/* List Buttons */}
      <button
        onClick={formatBulletList}
        className="px-3 py-1 rounded text-sm hover:bg-gray-100"
        aria-label="Bullet List"
      >
        • List
      </button>

      <button
        onClick={formatNumberedList}
        className="px-3 py-1 rounded text-sm hover:bg-gray-100"
        aria-label="Numbered List"
      >
        1. List
      </button>

      <div className="w-px h-6 bg-gray-300 mx-1" />

      {/* Alignment Buttons */}
      <button
        onClick={() => {
          editor.dispatchCommand(FORMAT_ELEMENT_COMMAND, 'left');
        }}
        className="px-3 py-1 rounded text-sm hover:bg-gray-100"
        aria-label="Align Left"
      >
        ⬅
      </button>

      <button
        onClick={() => {
          editor.dispatchCommand(FORMAT_ELEMENT_COMMAND, 'center');
        }}
        className="px-3 py-1 rounded text-sm hover:bg-gray-100"
        aria-label="Align Center"
      >
        ↔
      </button>

      <button
        onClick={() => {
          editor.dispatchCommand(FORMAT_ELEMENT_COMMAND, 'right');
        }}
        className="px-3 py-1 rounded text-sm hover:bg-gray-100"
        aria-label="Align Right"
      >
        ➡
      </button>
    </div>
  );
}
