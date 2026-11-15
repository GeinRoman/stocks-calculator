import type { ApiClient, AuthResponse } from './client.js';
import type { AllInfo, Profile } from './contracts.js';
import { portfolioApi } from './portfolio-api.js';

// Initial session and explicit profile switching load the complete server response.
// Individual mutations intentionally do not trigger /getall.
export function sessionApi(client: ApiClient) {
	const portfolio = portfolioApi(client);
	return {
		async login(user: string, pass: string): Promise<{ auth: AuthResponse; data: AllInfo }> {
			const auth = await client.authenticate('/login', user, pass);
			return { auth, data: await portfolio.getAll() };
		},
		async register(user: string, pass: string): Promise<{ auth: AuthResponse; data: AllInfo }> {
			const auth = await client.authenticate('/createuser', user, pass);
			return { auth, data: await portfolio.getAll() };
		},
		async restore(): Promise<AllInfo> {
			await client.refresh();
			return portfolio.getAll();
		},
		async selectProfile(profile: Profile): Promise<AllInfo> {
			await portfolio.profiles.setDefault(profile);
			return portfolio.getAll();
		}
	};
}
