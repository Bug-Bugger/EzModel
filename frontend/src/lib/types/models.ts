export interface User {
	id: string;
	email: string;
	username: string;
}

export interface Project {
	id: string;
	name: string;
	description: string;
	owner_id: string;
	database_type: 'postgresql' | 'mysql' | 'sqlite' | 'sqlserver';
	canvas_data: string;
	created_at: string;
	updated_at: string;
	owner?: User;
	collaborators?: User[];
}

export interface CreateProjectRequest {
	name: string;
	description?: string;
	database_type: 'postgresql' | 'mysql' | 'sqlite' | 'sqlserver';
}

export interface UpdateProjectRequest {
	name?: string;
	description?: string;
	database_type?: 'postgresql' | 'mysql' | 'sqlite' | 'sqlserver';
}