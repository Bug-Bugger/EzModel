import { writable } from 'svelte/store';
import type { Project, User } from '$lib/types/models';
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

	// Auto-save debounce timer
	let autoSaveTimeoutId: ReturnType<typeof setTimeout> | null = null;
	const DEBOUNCE_DELAY = 1000; // 1 second

	const store = {
		subscribe,

		// Load user's projects
		async loadProjects() {
			update((state) => ({ ...state, isLoading: true }));
			try {
				const projects = await projectService.getMyProjects();
				update((state) => ({ ...state, projects, isLoading: false }));
			} catch (error) {
				console.error('Failed to load projects:', error);
				update((state) => ({ ...state, isLoading: false }));
				throw error;
			}
		},

		// Set current project
		async setCurrentProject(projectId: string) {
			update((state) => ({ ...state, isLoading: true }));
			try {
				const project = await projectService.getProject(projectId);
				update((state) => ({ ...state, currentProject: project, isLoading: false }));
				return project;
			} catch (error) {
				console.error('Failed to load project:', error);
				update((state) => ({ ...state, isLoading: false }));
				throw error;
			}
		},

		// Load project (alias for setCurrentProject for consistency)
		async loadProject(projectId: string) {
			return await this.setCurrentProject(projectId);
		},

		// Add new project to list
		addProject(project: Project) {
			update((state) => ({
				...state,
				projects: [project, ...state.projects]
			}));
		},

		// Update project in list
		updateProject(updatedProject: Project) {
			update((state) => ({
				...state,
				projects: state.projects.map((p) => (p.id === updatedProject.id ? updatedProject : p)),
				currentProject:
					state.currentProject?.id === updatedProject.id ? updatedProject : state.currentProject
			}));
		},

		// Remove project from list
		removeProject(projectId: string) {
			update((state) => ({
				...state,
				projects: state.projects.filter((p) => p.id !== projectId),
				currentProject: state.currentProject?.id === projectId ? null : state.currentProject
			}));
		},

		// Add collaborator to current project
		async addCollaborator(collaboratorId: string) {
			const currentProject = store.getCurrentProject();
			if (!currentProject) {
				throw new Error('No current project');
			}

			await projectService.addCollaborator(currentProject.id, collaboratorId);
			// Refresh project to get updated collaborators list
			await store.setCurrentProject(currentProject.id);
		},

		// Remove collaborator from current project
		async removeCollaborator(collaboratorId: string) {
			const currentProject = store.getCurrentProject();
			if (!currentProject) {
				throw new Error('No current project');
			}

			await projectService.removeCollaborator(currentProject.id, collaboratorId);
			// Refresh project to get updated collaborators list
			await store.setCurrentProject(currentProject.id);
		},

		// Get current project (helper method)
		getCurrentProject(): Project | null {
			let current: Project | null = null;
			update((state) => {
				current = state.currentProject;
				return state;
			});
			return current;
		},

		// Save canvas data to backend
		async saveCanvasData(canvasData: string): Promise<void> {
			const currentProject = store.getCurrentProject();
			if (!currentProject) {
				throw new Error('No current project to save canvas data');
			}

			try {
				await projectService.updateProjectCanvasData(currentProject.id, canvasData);

				// Update local project state
				update((state) => ({
					...state,
					currentProject: state.currentProject
						? { ...state.currentProject, canvas_data: canvasData }
						: null
				}));
			} catch (error) {
				console.error('Failed to save canvas data:', error);
				throw error;
			}
		},

		// Auto-save canvas data with debouncing
		autoSaveCanvasData(canvasData: string) {
			console.log('PROJECT STORE: autoSaveCanvasData called with data length:', canvasData.length);

			if (autoSaveTimeoutId) {
				console.log('PROJECT STORE: Clearing existing timeout');
				clearTimeout(autoSaveTimeoutId);
			}

			autoSaveTimeoutId = setTimeout(async () => {
				try {
					console.log('PROJECT STORE: Starting auto-save after debounce delay');
					await store.saveCanvasData(canvasData);
					console.log('PROJECT STORE: Auto-save completed successfully');
				} catch (error) {
					console.error('PROJECT STORE: Auto-save failed:', error);
				}
			}, DEBOUNCE_DELAY);

			console.log('PROJECT STORE: Auto-save scheduled with', DEBOUNCE_DELAY, 'ms delay');
		},

		// Clear all state
		clear() {
			set(initialState);
		},

		// Cleanup method to clear pending timeouts
		cleanup() {
			if (autoSaveTimeoutId) {
				clearTimeout(autoSaveTimeoutId);
				autoSaveTimeoutId = null;
			}
		}
	};

	return store;
}

export const projectStore = createProjectStore();
