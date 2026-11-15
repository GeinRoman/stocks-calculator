// TEMP: Demo authentication only. This is not a real credential or refresh token.
export const DEMO_REFRESH_KEY = 'stocks-calculator-demo-refresh';
const DEMO_REFRESH_VALUE = 'demo-session-user-v1';

export interface DemoStorage { getItem(key: string): string | null; setItem(key: string, value: string): void; removeItem(key: string): void }

export function validateDemoCredentials(user: string, pass: string): string | null {
	if (!user || !pass) return 'Введите имя пользователя и пароль.';
	if (new TextEncoder().encode(pass).length > 64) return 'Пароль не должен превышать 64 байта UTF-8.';
	if (user !== 'user' || pass !== '123') return 'Неверное имя пользователя или пароль.';
	return null;
}

export async function demoAuthenticate(user: string, pass: string, storage: DemoStorage): Promise<void> {
	const error = validateDemoCredentials(user, pass);
	if (error) throw new Error(error);
	try { storage.setItem(DEMO_REFRESH_KEY, DEMO_REFRESH_VALUE); }
	catch { throw new Error('Не удалось сохранить демонстрационную сессию в браузере.'); }
}

// TEMP: Simulates POST /reftoken locally; never contacts the backend.
export async function demoRefresh(storage: DemoStorage): Promise<boolean> {
	const token = storage.getItem(DEMO_REFRESH_KEY);
	if (!token) return false;
	if (token === DEMO_REFRESH_VALUE) return true;
	storage.removeItem(DEMO_REFRESH_KEY);
	return false;
}

export function demoLogout(storage: DemoStorage): void { storage.removeItem(DEMO_REFRESH_KEY); }
