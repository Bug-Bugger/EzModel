export interface User {
	id: string;
	email: string;
	username: string;
}

export interface Table {
	table_id: string;
	name: string;
	project_id: string;
	pos_x: number;
	pos_y: number;
	fields?: Field[];
	created_at: string;
	updated_at: string;
}

export interface Field {
	field_id: string;
	table_id: string;
	name: string;
	data_type: string;
	is_primary_key: boolean;
	is_nullable: boolean;
	default_value: string;
	position: number;
	created_at: string;
	updated_at: string;
}

export interface Relationship {
	relationship_id: string;
	project_id: string;
	source_table_id: string;
	source_field_id: string;
	target_table_id: string;
	target_field_id: string;
	relation_type: 'one_to_one' | 'one_to_many' | 'many_to_many';
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

export interface CreateTableRequest {
	name: string;
	pos_x: number;
	pos_y: number;
}

export interface UpdateTableRequest {
	name?: string;
	pos_x?: number;
	pos_y?: number;
}

export interface UpdateTablePositionRequest {
	pos_x: number;
	pos_y: number;
}

export interface CreateRelationshipRequest {
	source_table_id: string;
	source_field_id: string;
	target_table_id: string;
	target_field_id: string;
	relation_type: 'one_to_one' | 'one_to_many' | 'many_to_many';
}

export interface UpdateRelationshipRequest {
	source_table_id?: string;
	source_field_id?: string;
	target_table_id?: string;
	target_field_id?: string;
	relation_type?: 'one_to_one' | 'one_to_many' | 'many_to_many';
}

export interface CreateFieldRequest {
	name: string;
	data_type: string;
	is_primary_key: boolean;
	is_nullable: boolean;
	default_value?: string;
	position?: number;
}

export interface UpdateFieldRequest {
	name?: string;
	data_type?: string;
	is_primary_key?: boolean;
	is_nullable?: boolean;
	default_value?: string;
	position?: number;
}
