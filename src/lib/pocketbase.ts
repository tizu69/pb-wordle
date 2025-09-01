import PocketBase, { RecordService } from 'pocketbase';

type CollID<Target extends string = ''> = string;
export interface CollCategory {
	id: CollID;
	name: string;
	mustHint: boolean;
	mustPresent: boolean;
}
export interface CollWord {
	id: CollID;
	word: string;
	category: CollID<'categories'>;
	hint?: string;
	hintLong?: string;
}

interface PB extends PocketBase {
	collection(idOrName: 'categories'): RecordService<CollCategory>;
	collection(idOrName: 'words'): RecordService<CollWord>;
	collection(idOrName: string): RecordService;
}

const pb = new PocketBase('/') as PB;
export default pb;
