import type {IAbstract} from './IAbstract'
import type {IUser} from './IUser'

export interface ITaskActivityChange {
	old: unknown
	new: unknown
}

export interface ITaskActivity extends IAbstract {
	id: number
	taskId: number
	userId: number
	action: string
	changedFields: Record<string, ITaskActivityChange>
	createdAt: Date | string | null
	user?: IUser | null
}
