# Requirements Document

## Introduction

This feature extends OrgMind to support multi-tenant graph management, allowing each user to create and manage multiple isolated knowledge graphs. Currently, the system creates a single graph per user. With this enhancement, users will be able to create multiple graphs for different projects, topics, or use cases, with each graph maintaining its own isolated memory and knowledge structure in Zep Cloud.

## Glossary

- **OrgMind_System**: The complete document processing platform including frontend, backend, and integrations
- **Graph**: A distinct knowledge graph entity stored in Zep Cloud, identified by a unique Zep graph ID
- **Graph_Membership**: A many-to-many relationship record associating a user with a graph, enabling multi-user collaboration
- **Graph_Creator**: The user who originally created a graph, stored as creator_id in the graphs table
- **Graph_Management_UI**: The frontend interface for viewing, creating, updating, and deleting graphs
- **Graph_Service**: The backend service handling graph CRUD operations, membership management, and Zep graph lifecycle management
- **Graph_Repository**: The data access layer for graph metadata and membership records stored in PostgreSQL
- **Zep_Service**: The service integrating with Zep Cloud SDK for knowledge graph operations
- **Graph_Context**: The selected graph that documents and memories are added to
- **JWT_Token**: JSON Web Token used for authenticated API requests
- **User_ID**: The unique identifier for an authenticated user
- **Zep_Graph_ID**: The unique identifier assigned by Zep Cloud for a graph instance

## Requirements

### Requirement 1

**User Story:** As an authenticated user, I want to view a list of all my graphs when I log in, so that I can see and manage my different knowledge graphs

#### Acceptance Criteria

1. WHEN an authenticated user logs in successfully, THE OrgMind_System SHALL redirect the user to a graphs list page
2. WHEN the graphs list page loads, THE Graph_Management_UI SHALL fetch all graphs associated with the User_ID from the JWT_Token
3. WHEN the backend receives a request to list graphs, THE Graph_Service SHALL query the Graph_Repository for all graphs belonging to the User_ID
4. WHEN graphs exist for the user, THE Graph_Management_UI SHALL display each graph with its name, description, creation date, and document count
5. WHERE no graphs exist for the user, THE Graph_Management_UI SHALL display a message prompting the user to create their first graph

### Requirement 2

**User Story:** As an authenticated user, I want to create a new graph with a name and description, so that I can organize my documents into separate knowledge domains

#### Acceptance Criteria

1. WHEN the user clicks the create graph button, THE Graph_Management_UI SHALL display a form with fields for graph name and optional description
2. WHEN the user submits the create graph form with valid data, THE OrgMind_System SHALL send the graph details to the backend with the JWT_Token
3. WHEN the backend receives a create graph request, THE Graph_Service SHALL create a new graph in Zep Cloud using the Zep_Service
4. WHEN the Zep Cloud graph creation succeeds, THE Graph_Service SHALL store the graph metadata including Zep_Graph_ID, name, description, and User_ID in the Graph_Repository
5. WHEN the graph is created successfully, THE Graph_Management_UI SHALL refresh the graphs list and display the new graph

### Requirement 3

**User Story:** As an authenticated user, I want to update a graph's name and description, so that I can keep my graph metadata current and meaningful

#### Acceptance Criteria

1. WHEN the user clicks the edit button on a graph, THE Graph_Management_UI SHALL display an edit form pre-populated with the current name and description
2. WHEN the user submits the edit form with valid data, THE OrgMind_System SHALL send the updated graph details to the backend with the JWT_Token
3. WHEN the backend receives an update graph request, THE Graph_Service SHALL verify the graph belongs to the User_ID from the JWT_Token
4. IF the graph does not belong to the user, THEN THE Graph_Service SHALL return a forbidden error
5. WHEN the graph ownership is verified, THE Graph_Service SHALL update the graph metadata in the Graph_Repository

### Requirement 4

**User Story:** As an authenticated user, I want to delete a graph, so that I can remove graphs I no longer need

#### Acceptance Criteria

1. WHEN the user clicks the delete button on a graph, THE Graph_Management_UI SHALL display a confirmation dialog warning that deletion is permanent
2. WHEN the user confirms deletion, THE OrgMind_System SHALL send a delete request to the backend with the graph ID and JWT_Token
3. WHEN the backend receives a delete graph request, THE Graph_Service SHALL verify the graph belongs to the User_ID from the JWT_Token
4. IF the graph does not belong to the user, THEN THE Graph_Service SHALL return a forbidden error
5. WHEN the graph ownership is verified, THE Graph_Service SHALL delete the graph from Zep Cloud using the Zep_Service and remove the metadata from the Graph_Repository

### Requirement 5

**User Story:** As an authenticated user, I want to click on a graph to view its details and documents, so that I can see what content exists in that knowledge domain

#### Acceptance Criteria

1. WHEN the user clicks on a graph in the list, THE Graph_Management_UI SHALL navigate to the graph detail page with the graph ID in the URL
2. WHEN the graph detail page loads, THE OrgMind_System SHALL fetch the graph details including name, description, and metadata using the graph ID from the URL
3. WHEN the graph detail page renders, THE Graph_Management_UI SHALL display the graph name and description at the top of the page
4. WHEN the graph detail page renders, THE Graph_Management_UI SHALL display a section with a list of all documents associated with the graph from the documents database table
5. WHEN displaying documents, THE Graph_Management_UI SHALL show each document's type indicating whether it is user-created content or an uploaded file

### Requirement 6

**User Story:** As an authenticated user, I want to add a new document to a graph by choosing between editor or file upload, so that I can contribute content in my preferred format

#### Acceptance Criteria

1. WHEN the user clicks the add document button on the graph detail page, THE Graph_Management_UI SHALL display a modal with two options: edit document or upload file
2. WHEN the user selects edit document from the modal, THE Graph_Management_UI SHALL navigate to the Lexical editor component page with the graph ID as context
3. WHEN the user selects upload file from the modal, THE Graph_Management_UI SHALL navigate to the file upload component page with the graph ID as context
4. WHEN the user submits content from the Lexical editor, THE OrgMind_System SHALL include the graph ID in the document creation request
5. WHEN the user uploads a file, THE OrgMind_System SHALL include the graph ID in the file upload request

### Requirement 7

**User Story:** As an authenticated user, I want to create or upload documents that are associated with a specific graph, so that the knowledge is organized in the correct context

#### Acceptance Criteria

1. WHEN the backend receives a document creation request with a graph ID, THE Graph_Service SHALL verify the graph belongs to the User_ID from the JWT_Token
2. IF the graph does not belong to the user, THEN THE Graph_Service SHALL return a forbidden error
3. WHEN the graph ownership is verified, THE Processing_Service SHALL chunk and process the document content
4. WHEN the document is processed, THE Zep_Service SHALL add the document chunks to the specific Zep_Graph_ID as memory entries
5. WHEN the document is successfully created, THE Graph_Repository SHALL store the document record with the graph ID, document type, and User_ID

### Requirement 8

**User Story:** As an authenticated user, I want to view the list of documents in a graph, so that I can see what content has been added and manage individual documents

#### Acceptance Criteria

1. WHEN the graph detail page loads, THE Graph_Management_UI SHALL fetch all documents associated with the graph ID from the backend
2. WHEN the backend receives a request to list graph documents, THE Graph_Service SHALL verify the graph belongs to the User_ID from the JWT_Token
3. IF the graph does not belong to the user, THEN THE Graph_Service SHALL return a forbidden error
4. WHEN the graph ownership is verified, THE Graph_Repository SHALL return all documents associated with the graph ID
5. WHEN displaying documents, THE Graph_Management_UI SHALL show each document's name, type (user-created or uploaded file), creation date, and size

### Requirement 9

**User Story:** As an authenticated user, I want to view the knowledge graph visualization for a specific graph, so that I can understand the relationships in that graph's context

#### Acceptance Criteria

1. WHEN the user views a graph detail page, THE Graph_Management_UI SHALL fetch the knowledge graph data for the specific Zep_Graph_ID
2. WHEN the backend receives a graph visualization request, THE Graph_Service SHALL verify the graph belongs to the User_ID from the JWT_Token
3. IF the graph does not belong to the user, THEN THE Graph_Service SHALL return a forbidden error
4. WHEN the graph ownership is verified, THE Zep_Service SHALL query Zep Cloud for the knowledge graph data using the Zep_Graph_ID
5. WHEN the graph data is retrieved, THE Graph_Management_UI SHALL render the knowledge graph using the sigma.js visualizer

### Requirement 10

**User Story:** As the system, I want to store graph metadata in the database, so that users can manage multiple graphs with persistent metadata

#### Acceptance Criteria

1. WHEN a graph is created, THE Graph_Repository SHALL store the graph record with User_ID, Zep_Graph_ID, name, description, and timestamps
2. WHEN storing a graph, THE Graph_Repository SHALL create a unique database ID for the graph record
3. WHEN a document is added to a graph, THE Graph_Repository SHALL update the document count for that graph
4. WHEN a graph is deleted, THE Graph_Repository SHALL cascade delete all associated documents from the database
5. WHEN querying graphs, THE Graph_Repository SHALL only return graphs belonging to the requesting User_ID

### Requirement 11

**User Story:** As the system, I want to refactor the Zep service to use graph IDs, so that memories and knowledge graphs are properly isolated per graph

#### Acceptance Criteria

1. WHEN adding memory to Zep Cloud, THE Zep_Service SHALL use the Zep_Graph_ID parameter to target the specific graph
2. WHEN retrieving graph data from Zep Cloud, THE Zep_Service SHALL use the Zep_Graph_ID parameter to fetch the specific graph
3. WHEN creating a new graph in Zep Cloud, THE Zep_Service SHALL return the Zep_Graph_ID assigned by Zep Cloud
4. WHEN deleting a graph from Zep Cloud, THE Zep_Service SHALL use the Zep_Graph_ID to target the specific graph for deletion
5. IF a Zep Cloud operation fails, THEN THE Zep_Service SHALL return an error with context about which graph operation failed

### Requirement 12

**User Story:** As the system, I want to maintain backward compatibility with existing APIs, so that the refactoring does not break current functionality

#### Acceptance Criteria

1. THE OrgMind_System SHALL preserve all existing API endpoints for authentication, document operations, and graph visualization
2. THE OrgMind_System SHALL add new API endpoints for graph CRUD operations without modifying existing endpoints
3. WHERE existing code references a single user graph, THE OrgMind_System SHALL refactor to use a graph context parameter
4. THE OrgMind_System SHALL maintain all existing frontend components including Lexical editor, file upload, and graph visualizer
5. THE OrgMind_System SHALL update the home page route to redirect to the graphs list page instead of the editor

### Requirement 13

**User Story:** As an authenticated user, I want to see how many documents are in each graph, so that I can understand the content volume in each knowledge domain

#### Acceptance Criteria

1. WHEN the graphs list page displays graphs, THE Graph_Management_UI SHALL show the document count for each graph
2. WHEN a document is successfully added to a graph, THE Graph_Service SHALL increment the document count for that graph
3. WHEN a document is deleted from a graph, THE Graph_Service SHALL decrement the document count for that graph
4. WHEN the graph detail page loads, THE Graph_Management_UI SHALL display the current document count for the graph
5. WHEN no documents exist in a graph, THE Graph_Management_UI SHALL display zero as the document count

### Requirement 14

**User Story:** As the system, I want to manage graph membership through a many-to-many relationship, so that multiple users can access and collaborate on the same graph

#### Acceptance Criteria

1. WHEN a graph is created, THE Graph_Service SHALL create a graph membership record associating the creator User_ID with the graph as the owner
2. WHEN a user is added to a graph, THE Graph_Service SHALL create a graph membership record with the user's User_ID and the graph ID
3. WHEN a user requests to list graphs, THE Graph_Repository SHALL query the graph_memberships table to return all graphs the user is a member of
4. WHEN a user attempts to access a graph, THE Graph_Service SHALL verify the user has a membership record for that graph
5. IF the user does not have a membership record for the graph, THEN THE Graph_Service SHALL return a forbidden error

### Requirement 15

**User Story:** As an authenticated user, I want to see only the graphs I am a member of, so that I can access graphs I created or was invited to

#### Acceptance Criteria

1. WHEN the user logs in and navigates to the graphs list page, THE OrgMind_System SHALL fetch all graphs where the user has a membership record
2. WHEN displaying graphs, THE Graph_Management_UI SHALL show graphs created by the user and graphs the user was added to
3. WHEN the backend receives a list graphs request, THE Graph_Repository SHALL join the graphs and graph_memberships tables filtered by User_ID
4. WHEN a graph is deleted by its creator, THE Graph_Service SHALL delete all associated graph membership records
5. WHEN listing graphs, THE Graph_Management_UI SHALL indicate which graphs the user created versus which they are a member of

### Requirement 16

**User Story:** As an authenticated user, I want to click on a document in the graph detail page, so that I can view or edit its content

#### Acceptance Criteria

1. WHEN the user clicks on a document in the documents list, THE Graph_Management_UI SHALL navigate to the document view page
2. WHERE the document type is user-created content, THE Graph_Management_UI SHALL display the content in the Lexical editor for viewing or editing
3. WHERE the document type is an uploaded file, THE Graph_Management_UI SHALL display the file metadata and provide a download option
4. WHEN the user edits a user-created document, THE OrgMind_System SHALL save the changes and re-process the content through the Zep_Service
5. WHEN the backend receives a document update request, THE Graph_Service SHALL verify the document belongs to a graph the user is a member of
