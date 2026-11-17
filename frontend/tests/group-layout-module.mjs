import { readFile } from 'node:fs/promises';
import ts from 'typescript';

const source = await readFile(new URL('../src/lib/logic/tabs/group-layout.ts', import.meta.url), 'utf8');
const { outputText } = ts.transpileModule(source, {
	compilerOptions: { target: ts.ScriptTarget.ES2022, module: ts.ModuleKind.ESNext }
});
export const groupLayoutModuleUrl = `data:text/javascript;base64,${Buffer.from(outputText).toString('base64')}`;
