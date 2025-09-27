export interface ApiResponse<T = any> {
	success: boolean;
	message: string;
	data?: T;
}

export interface LoginRequest {
	email: string;
	password: string;
}

export interface RegisterRequest {
	email: string;
	username: string;
	password: string;
}

export interface LoginResponse {
	access_token: string;
	refresh_token: string;
	token_type: string;
	expires_in: number;
}

export interface RefreshTokenRequest {
	refresh_token: string;
}

export interface UpdateUserRequest {
	email?: string;
	username?: string;
}

export interface UpdatePasswordRequest {
	current_password: string;
	new_password: string;
}

export interface ApiError {
	message: string;
	status: number;
	code?: string;
}