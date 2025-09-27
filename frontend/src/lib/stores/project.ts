import { writable } from 'svelte/store';
import type { Project } from '$lib/types/models';
import { projectService } from '$lib/services/project';

interface ProjectState {
	projects: Project[];
	currentProject: Project | null;
	isLoading: boolean;
}

function createProjectStore() {
	const initialState: ProjectState = {
		projects: [],
		currentProject: null,
		isLoading: false
	};

	const { subscribe, set, update } = writable(initialState);

	return {
		subscribe,

		// Load user's projects
		async loadProjects() {
			update(state => ({ ...state, isLoading: true }));
			try {
				const projects = await projectService.getMyProjects();
				update(state => ({ ...state, projects, isLoading: false }));
			} catch (error) {
				console.error('Failed to load projects:', error);
				update(state => ({ ...state, isLoading: false }));
				throw error;
			}
		},

		// Set current project
		async setCurrentProject(projectId: string) {
			update(state => ({ ...state, isLoading: true }));
			try {
				const project = await projectService.getProject(projectId);
				update(state => ({ ...state, currentProject: project, isLoading: false }));
				return project;
			} catch (error) {
				console.error('Failed to load project:', error);
				update(state => ({ ...state, isLoading: false }));
				throw error;
			}
		},

		// Load project (alias for setCurrentProject for consistency)
		async loadProject(projectId: string) {
			return await this.setCurrentProject(projectId);
		},

		// Add new project to list
		addProject(project: Project) {
			update(state => ({
				...state,
				projects: [project, ...state.projects]
			}));
		},

		// Update project in list
		updateProject(updatedProject: Project) {
			update(state => ({
				...state,
				projects: state.projects.map(p =>
					p.id === updatedProject.id ? updatedProject : p
				),
				currentProject: state.currentProject?.id === updatedProject.id
					? updatedProject
					: state.currentProject
			}));
		},

		// Remove project from list
		removeProject(projectId: string) {
			update(state => ({
				...state,
				projects: state.projects.filter(p => p.id !== projectId),
				currentProject: state.currentProject?.id === projectId
					? null
					: state.currentProject
			}));
		},

		// Clear all state
		clear() {
			set(initialState);
		}
	};
}

export const projectStore = createProjectStore();