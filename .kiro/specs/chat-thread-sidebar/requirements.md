# Requirements Document

## Introduction

This feature redesigns the chat interface to include a thread list sidebar, enabling users to manage multiple conversation threads within a graph. The current implementation auto-creates a thread on load, but the new design will display a list of existing threads and allow users to create new threads by typing their first message.

## Glossary

- **ChatInterface**: The main React component that orchestrates the chat experience
- **ThreadList**: A new component that displays all chat threads for a graph
- **Thread**: A conversation session containing multiple messages between user and AI
- **Graph**: A knowledge graph context within which chat threads exist
- **Backend API**: The Go-based REST API that manages threads and messages

## Requirements

### Requirement 1

**User Story:** As a user, I want to see a list of my existing chat threads for a graph, so that I can resume previous conversations

#### Acceptance Criteria

1. WHEN THE ChatInterface loads, THE ChatInterface SHALL display a ThreadList component in the left column
2. THE ThreadList SHALL fetch all threads for the current graph from the backend API
3. THE ThreadList SHALL display each thread with a preview of its first message or creation timestamp
4. WHILE the threads are loading, THE ThreadList SHALL display a loading skeleton
5. IF the thread fetch fails, THEN THE ThreadList SHALL display an error message with retry option

### Requirement 2

**User Story:** As a user, I want to select a thread from the list, so that I can view and continue that conversation

#### Acceptance Criteria

1. WHEN a user clicks on a thread in the ThreadList, THE ChatInterface SHALL display that thread's messages in the right column
2. THE ChatInterface SHALL highlight the selected thread in the ThreadList
3. THE ChatInterface SHALL load the selected thread's message history from the backend
4. THE ChatInterface SHALL display the ChatMessageList and ChatInput components for the selected thread
5. WHILE messages are loading, THE ChatInterface SHALL display a loading state in the message area

### Requirement 3

**User Story:** As a user, I want to start a new conversation when no thread is selected, so that I can ask my first question without pre-selecting anything

#### Acceptance Criteria

1. WHEN no thread is selected, THE ChatInterface SHALL display a ChatInput-like component in the right column
2. THE ChatInterface SHALL display a placeholder or welcome message indicating the user can start a new conversation
3. WHEN the user types and sends their first message, THE ChatInterface SHALL call the backend API to create a new thread
4. WHEN the new thread is created, THE ChatInterface SHALL add it to the ThreadList and select it automatically
5. THE ChatInterface SHALL display the new thread's conversation in the right column with the user's message and AI response

### Requirement 4

**User Story:** As a user, I want the interface to be responsive and work well on different screen sizes, so that I can use it on desktop and mobile devices

#### Acceptance Criteria

1. THE ChatInterface SHALL use a two-column layout on desktop screens (width >= 768px)
2. THE ChatInterface SHALL collapse the ThreadList into a toggleable sidebar on mobile screens (width < 768px)
3. WHEN on mobile, THE ChatInterface SHALL show a menu button to toggle the ThreadList visibility
4. WHEN a thread is selected on mobile, THE ChatInterface SHALL automatically hide the ThreadList to show the conversation
5. THE ChatInterface SHALL maintain proper spacing and readability across all screen sizes

### Requirement 5

**User Story:** As a user, I want to see visual feedback when creating a new thread, so that I understand the system is processing my request

#### Acceptance Criteria

1. WHEN the user sends their first message to create a thread, THE ChatInterface SHALL disable the input and show a loading state
2. THE ChatInterface SHALL display the user's message immediately in the conversation area
3. WHILE the thread is being created, THE ChatInterface SHALL show a loading indicator in the ThreadList
4. WHEN the thread creation completes, THE ChatInterface SHALL update the ThreadList with the new thread
5. IF thread creation fails, THEN THE ChatInterface SHALL display an error message and allow the user to retry

### Requirement 6

**User Story:** As a user, I want the thread list to update when I create or interact with threads, so that I always see the current state

#### Acceptance Criteria

1. WHEN a new thread is created, THE ThreadList SHALL add it to the top of the list
2. WHEN a thread receives a new message, THE ThreadList SHALL update that thread's preview
3. THE ThreadList SHALL sort threads by most recent activity (newest first)
4. THE ThreadList SHALL persist the selected thread state when the component re-renders
5. THE ChatInterface SHALL maintain scroll position in the ThreadList when threads are updated
