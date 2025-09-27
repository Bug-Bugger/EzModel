export interface User {
	id: string;
	email: string;
	username: string;
	name: string;
}

export interface Table {
	id: string;
	name: string;
	project_id: string;
	fields?: Field[];
	created_at: string;
	updated_at: string;
}

export interface Field {
	id: string;
	name: string;
	type: string;
	table_id: string;
	is_primary: boolean;
	is_foreign: boolean;
	is_required: boolean;
	is_unique: boolean;
	default_value?: string;
	constraints?: string[];
	created_at: string;
	updated_at: string;
}

export interface Relationship {
	id: string;
	name: string;
	project_id: string;
	from_table_id: string;
	to_table_id: string;
	from_field_id: string;
	to_field_id: string;
	relationship_type: 'one-to-one' | 'one-to-many' | 'many-to-many';
	created_at: string;
	updated_at: string;
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
	tables?: Table[];
	relationships?: Relationship[];
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