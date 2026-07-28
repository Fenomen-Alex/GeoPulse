import {createSignal, createContext, useContext, type Component} from 'solid-js';

export interface User {
	id: string;
	name: string;
	email: string;
	avatar?: string;
	daily_quota: number;
}

const AuthContext = createContext<{
	user: () => User | null;
	login: () => Promise<void>;
	logout: () => Promise<void>;
}>();

export const useAuth = () => useContext(AuthContext);

export const AuthProvider: Component<{ children: any }> = (props) => {
	const [user, setUser] = createSignal<User | null>(null);

	const checkAuthStatus = async () => {
		try {
			const res = await fetch('/api/v1/auth/status');
			if (res.ok) {
				const data = await res.json();
				if (data.authenticated) {
					setUser(data.user);
				}
			}
		} catch (err) {
			console.error('Auth check failed:', err);
		}
	};

	checkAuthStatus();

	const login = async () => {
		try {
			window.location.href = '/api/v1/auth/login';
		} catch (err) {
			console.error('Login failed:', err);
		}
	};

	const logout = async () => {
		try {
			await fetch('/api/v1/auth/logout', { method: 'POST' });
			setUser(null);
		} catch (err) {
			console.error('Logout failed:', err);
		}
	};

	return (
		<AuthContext.Provider value={{ user, login, logout }}>
			{props.children}
		</AuthContext.Provider>
	);
};
