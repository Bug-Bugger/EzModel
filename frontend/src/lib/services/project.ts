import { apiClient } from './api';
import type { Project, CreateProjectRequest, UpdateProjectRequest } from '$lib/types/models';

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
}

export const projectService = new ProjectService();