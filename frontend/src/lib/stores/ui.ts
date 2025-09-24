import { writable } from 'svelte/store';

interface Toast {
	id: string;
	title: string;
	description?: string;
	type: 'success' | 'error' | 'warning' | 'info';
}

interface UIState {
	toasts: Toast[];
	isLoading: boolean;
}

function createUIStore() {
	const initialState: UIState = {
		toasts: [],
		isLoading: false
	};

	const { subscribe, update } = writable(initialState);

	return {
		subscribe,

		// Add toast notification
		addToast(toast: Omit<Toast, 'id'>) {
			const id = Math.random().toString(36).substr(2, 9);
			update(state => ({
				...state,
				toasts: [...state.toasts, { ...toast, id }]
			}));

			// Auto-remove toast after 5 seconds
			setTimeout(() => {
				update(state => ({
					...state,
					toasts: state.toasts.filter(t => t.id !== id)
				}));
			}, 5000);

			return id;
		},

		// Remove specific toast
		removeToast(id: string) {
			update(state => ({
				...state,
				toasts: state.toasts.filter(t => t.id !== id)
			}));
		},

		// Clear all toasts
		clearToasts() {
			update(state => ({ ...state, toasts: [] }));
		},

		// Set global loading state
		setLoading(isLoading: boolean) {
			update(state => ({ ...state, isLoading }));
		},

		// Show success toast
		success(title: string, description?: string) {
			return this.addToast({ title, description, type: 'success' });
		},

		// Show error toast
		error(title: string, description?: string) {
			return this.addToast({ title, description, type: 'error' });
		},

		// Show warning toast
		warning(title: string, description?: string) {
			return this.addToast({ title, description, type: 'warning' });
		},

		// Show info toast
		info(title: string, description?: string) {
			return this.addToast({ title, description, type: 'info' });
		}
	};
}

export const uiStore = createUIStore();