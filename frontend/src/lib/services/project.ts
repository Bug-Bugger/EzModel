import { apiClient } from './api';
import type {
	Project,
	CreateProjectRequest,
	UpdateProjectRequest,
	Table,
	CreateTableRequest,
	UpdateTableRequest,
	UpdateTablePositionRequest,
	Relationship,
	CreateRelationshipRequest,
	UpdateRelationshipRequest
} from '$lib/types/models';

export class ProjectService {
	async getMyProjects(): Promise<Project[]> {
		const response = await apiClient.get<Project[]>('/projects/my');
		if (response.success && response.data) {
			return response.data;
		}
		return [];
	}

	async getProject(id: string): Promise<Project> {
		const response = await apiClient.get<Project>(`/projects/${id}`);
		if (response.success && response.data) {
			return response.data;
		}
		throw new Error(response.message || 'Failed to fetch project');
	}

	async createProject(projectData: CreateProjectRequest): Promise<Project> {
		const response = await apiClient.post<Project>('/projects', projectData);
		if (response.success && response.data) {
			return response.data;
		}
		throw new Error(response.message || 'Failed to create project');
	}

	async updateProject(id: string, projectData: UpdateProjectRequest): Promise<Project> {
		const response = await apiClient.put<Project>(`/projects/${id}`, projectData);
		if (response.success && response.data) {
			return response.data;
		}
		throw new Error(response.message || 'Failed to update project');
	}

	async deleteProject(id: string): Promise<void> {
		const response = await apiClient.delete(`/projects/${id}`);
		if (!response.success) {
			throw new Error(response.message || 'Failed to delete project');
		}
	}

	async addCollaborator(projectId: string, collaboratorId: string): Promise<void> {
		const response = await apiClient.post(`/projects/${projectId}/collaborators`, {
			collaborator_id: collaboratorId
		});
		if (!response.success) {
			throw new Error(response.message || 'Failed to add collaborator');
		}
	}

	async removeCollaborator(projectId: string, collaboratorId: string): Promise<void> {
		const response = await apiClient.delete(`/projects/${projectId}/collaborators/${collaboratorId}`);
		if (!response.success) {
			throw new Error(response.message || 'Failed to remove collaborator');
		}
	}

	// Table Management Methods
	async createTable(projectId: string, tableData: CreateTableRequest): Promise<Table> {
		const response = await apiClient.post<Table>(`/projects/${projectId}/tables`, tableData);
		if (response.success && response.data) {
			return response.data;
		}
		throw new Error(response.message || 'Failed to create table');
	}

	async getProjectTables(projectId: string): Promise<Table[]> {
		const response = await apiClient.get<Table[]>(`/projects/${projectId}/tables`);
		if (response.success && response.data) {
			return response.data;
		}
		return [];
	}

	async getTable(projectId: string, tableId: string): Promise<Table> {
		const response = await apiClient.get<Table>(`/projects/${projectId}/tables/${tableId}`);
		if (response.success && response.data) {
			return response.data;
		}
		throw new Error(response.message || 'Failed to fetch table');
	}

	async updateTable(projectId: string, tableId: string, tableData: UpdateTableRequest): Promise<Table> {
		const response = await apiClient.put<Table>(`/projects/${projectId}/tables/${tableId}`, tableData);
		if (response.success && response.data) {
			return response.data;
		}
		throw new Error(response.message || 'Failed to update table');
	}

	async updateTablePosition(projectId: string, tableId: string, positionData: UpdateTablePositionRequest): Promise<void> {
		const response = await apiClient.put(`/projects/${projectId}/tables/${tableId}/position`, positionData);
		if (!response.success) {
			throw new Error(response.message || 'Failed to update table position');
		}
	}

	async deleteTable(projectId: string, tableId: string): Promise<void> {
		const response = await apiClient.delete(`/projects/${projectId}/tables/${tableId}`);
		if (!response.success) {
			throw new Error(response.message || 'Failed to delete table');
		}
	}

	async getTableFields(projectId: string, tableId: string): Promise<any[]> {
		const response = await apiClient.get<any[]>(`/projects/${projectId}/tables/${tableId}/fields`);
		if (response.success && response.data) {
			return response.data;
		}
		return [];
	}

	async updateProjectCanvasData(projectId: string, canvasData: string): Promise<void> {
		const response = await apiClient.put(`/projects/${projectId}`, { canvas_data: canvasData });
		if (!response.success) {
			throw new Error(response.message || 'Failed to update canvas data');
		}
	}

	// Relationship Management Methods
	async createRelationship(projectId: string, relationshipData: CreateRelationshipRequest): Promise<Relationship> {
		const response = await apiClient.post<Relationship>(`/projects/${projectId}/relationships`, relationshipData);
		if (response.success && response.data) {
			return response.data;
		}
		throw new Error(response.message || 'Failed to create relationship');
	}

	async getProjectRelationships(projectId: string): Promise<Relationship[]> {
		const response = await apiClient.get<Relationship[]>(`/projects/${projectId}/relationships`);
		if (response.success && response.data) {
			return response.data;
		}
		return [];
	}

	async getRelationship(projectId: string, relationshipId: string): Promise<Relationship> {
		const response = await apiClient.get<Relationship>(`/projects/${projectId}/relationships/${relationshipId}`);
		if (response.success && response.data) {
			return response.data;
		}
		throw new Error(response.message || 'Failed to fetch relationship');
	}

	async updateRelationship(projectId: string, relationshipId: string, relationshipData: UpdateRelationshipRequest): Promise<Relationship> {
		const response = await apiClient.put<Relationship>(`/projects/${projectId}/relationships/${relationshipId}`, relationshipData);
		if (response.success && response.data) {
			return response.data;
		}
		throw new Error(response.message || 'Failed to update relationship');
	}

	async deleteRelationship(projectId: string, relationshipId: string): Promise<void> {
		const response = await apiClient.delete(`/projects/${projectId}/relationships/${relationshipId}`);
		if (!response.success) {
			throw new Error(response.message || 'Failed to delete relationship');
		}
	}
}

export const projectService = new ProjectService();