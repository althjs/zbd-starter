/**
 * 웹하드 확장 자리 — 다른 기능이 웹하드에 「특별한 파일」(누르면 그 기능으로 열림)·파일 메뉴·「새로 만들기」 를 더한다.
 * 기능마다 ext/{기능}.ts 한 파일(웹오피스 = ext/office.ts · 포토샵 = ext/photo.ts) — 여기서 이름 순으로 모은다.
 * 스타터에는 ext/ 가 비어 있고, 플러그인이 자기 파일을 넣는다(이 파일은 건드리지 않는다).
 */
import type { ExplorerExtModule, ExplorerFileAction, ExplorerFileExt, ExplorerNewItem } from './ext-types.ts';
export type { ExplorerFileAction, ExplorerFileExt, ExplorerNewItem } from './ext-types.ts';

const MODS: ExplorerExtModule[] = Object.entries(import.meta.glob<ExplorerExtModule>('./ext/*.ts', { eager: true }))
	.sort(([a], [b]) => (a < b ? -1 : 1))
	.map(([, m]) => m);

export function fileExt(name: string): ExplorerFileExt | null {
	for (const m of MODS) {
		const x = m.fileExt?.(name);
		if (x) return x;
	}
	return null;
}

export function fileActions(name: string): ExplorerFileAction[] {
	return MODS.flatMap((m) => m.fileActions?.(name) || []);
}

export const newItems: ExplorerNewItem[] = MODS.flatMap((m) => m.newItems || []);
