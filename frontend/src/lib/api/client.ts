export interface TokenStorage {
	getItem(key: string): string | null;
	setItem(key: string, value: string): void;
	removeItem(key: string): void;
}

export interface AccessToken { access_token: string; expires_at: string }
export interface AuthResponse { token: AccessToken; ref_token: string; profiles: { name: string; default: boolean }[] }
export const REFRESH_TOKEN_KEY = 'stocks-calculator-refresh-token';

export class HttpError extends Error {
	constructor(public readonly status: number, message: string) { super(message || `HTTP ${status}`); this.name = 'HttpError'; }
}

export class SessionExpiredError extends Error {
	constructor() { super('Session expired'); this.name = 'SessionExpiredError'; }
}

type RequestOptions = { method?: 'GET' | 'POST' | 'PATCH' | 'DELETE'; body?: unknown; signal?: AbortSignal };
type ClientOptions = {
	baseUrl?: string;
	storage: TokenStorage;
	fetch?: typeof fetch;
	onSessionExpired?: () => void;
	now?: () => number;
};

// One client instance belongs to one browser session; do not share it between SSR requests.
export class ApiClient {
	private readonly baseUrl: string;
	private readonly storage: TokenStorage;
	private readonly fetcher: typeof fetch;
	private readonly onSessionExpired?: () => void;
	private readonly now: () => number;
	private token: AccessToken | null = null;
	private generation = 0;
	private refreshPromise: Promise<void> | null = null;

	constructor(options: ClientOptions) {
		this.baseUrl = (options.baseUrl ?? '').replace(/\/$/, '');
		this.storage = options.storage;
		this.fetcher = options.fetch ?? globalThis.fetch.bind(globalThis);
		this.onSessionExpired = options.onSessionExpired;
		this.now = options.now ?? Date.now;
	}

	private async send(path: string, options: RequestOptions, token?: string): Promise<Response> {
		const headers = new Headers();
		if (token) headers.set('Authorization', `Bearer ${token}`);
		if (options.body !== undefined) headers.set('Content-Type', 'application/json');
		return this.fetcher(`${this.baseUrl}${path}`, {
			method: options.method ?? 'GET', headers, signal: options.signal,
			body: options.body === undefined ? undefined : JSON.stringify(options.body)
		});
	}

	private async ensureOk(response: Response): Promise<void> {
		if (response.status !== 200) throw new HttpError(response.status, (await response.text()).trim());
	}

	private validToken(value: AccessToken): boolean {
		return typeof value?.access_token === 'string' && value.access_token.length > 0 &&
			typeof value.expires_at === 'string' && Number.isFinite(Date.parse(value.expires_at)) && Date.parse(value.expires_at) > this.now();
	}

	async authenticate(path: '/login' | '/createuser', user: string, pass: string): Promise<AuthResponse> {
		const generation = this.generation;
		const response = await this.send(path, { method: 'POST', body: { user, pass } });
		await this.ensureOk(response);
		const data = await response.json() as AuthResponse;
		if (!this.validToken(data?.token) || typeof data.ref_token !== 'string' || !data.ref_token || !Array.isArray(data.profiles)) throw new Error('Invalid authentication response');
		if (generation !== this.generation) throw new SessionExpiredError();
		this.storage.setItem(REFRESH_TOKEN_KEY, data.ref_token);
		this.generation++;
		this.refreshPromise = null;
		this.token = { ...data.token };
		return data;
	}

	logout(): void {
		this.generation++;
		this.token = null;
		this.refreshPromise = null;
		this.storage.removeItem(REFRESH_TOKEN_KEY);
	}

	private expire(generation: number): void {
		if (generation !== this.generation) return;
		this.generation++;
		this.token = null;
		this.refreshPromise = null;
		try { this.storage.removeItem(REFRESH_TOKEN_KEY); } finally { this.onSessionExpired?.(); }
	}

	async refresh(): Promise<void> {
		if (this.refreshPromise) return this.refreshPromise;
		const generation = this.generation;
		const promise = (async () => {
			try {
				const refreshToken = this.storage.getItem(REFRESH_TOKEN_KEY);
				if (!refreshToken) throw new SessionExpiredError();
				const response = await this.send('/reftoken', { method: 'POST', body: refreshToken });
				await this.ensureOk(response);
				const token = await response.json() as AccessToken;
				if (!this.validToken(token)) throw new Error('Invalid refresh response');
				if (generation !== this.generation) throw new SessionExpiredError();
				this.token = token;
			} catch (error) {
				this.expire(generation);
				throw error;
			}
		})();
		this.refreshPromise = promise;
		try { await promise; } finally { if (this.refreshPromise === promise) this.refreshPromise = null; }
	}

	async request<T>(path: string, options: RequestOptions = {}): Promise<T> {
		if (!this.token || !this.validToken(this.token)) await this.refresh();
		const generation = this.generation;
		const usedToken = this.token!.access_token;
		let response = await this.send(path, options, usedToken);
		if (response.status === 401) {
			if (generation !== this.generation) throw new SessionExpiredError();
			if (this.token?.access_token === usedToken) await this.refresh();
			response = await this.send(path, options, this.token!.access_token);
			if (response.status === 401) { this.expire(generation); throw new SessionExpiredError(); }
		}
		if (generation !== this.generation) throw new SessionExpiredError();
		await this.ensureOk(response);
		const text = await response.text();
		return (text ? JSON.parse(text) : undefined) as T;
	}
}
