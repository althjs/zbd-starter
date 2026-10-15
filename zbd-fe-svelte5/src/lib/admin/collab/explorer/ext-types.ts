/** 웹하드 확장 모양 — ext.ts 가 ext/*.ts(기능마다 한 파일 — 웹오피스·포토샵 …)를 모아 웹하드에 건넨다 */
import type { FILE_NODE } from '#lib/shared/types/file.ts';

/** 파일마다 더하는 일(오른쪽 누르기 메뉴·미리보기 아래 단추) — 특별한 파일이 아니어도(예: 그림 → 포토샵에서 편집) */
export type ExplorerFileAction = { icon: string; label: string; run: (root: string, node: FILE_NODE) => void };
/** 특별한 파일 — 이름표·아이콘·색, 누르면 여는 일 */
export type ExplorerFileExt = { label: string; icon: string; color: string; openLabel: string; hint: string; open: (root: string, node: FILE_NODE) => Promise<void> };
/** 폴더 메뉴의 「새 …」 — 만든 파일 경로를 돌려주면 웹하드가 목록을 다시 읽어 고른다 */
export type ExplorerNewItem = { icon: string; label: string; create: (root: string, folder: string) => Promise<string | null> };
/** ext/*.ts 한 파일이 내보내는 것(모두 없어도 된다) */
export type ExplorerExtModule = { fileExt?: (name: string) => ExplorerFileExt | null; fileActions?: (name: string) => ExplorerFileAction[]; newItems?: ExplorerNewItem[] };
