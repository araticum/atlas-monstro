import AbstractModel from './abstractModel'
import UserModel from './user'
import type {ITaskActivity} from '@/modelTypes/ITaskActivity'
import type {IUser} from '@/modelTypes/IUser'

export default class TaskActivityModel extends AbstractModel<ITaskActivity> implements ITaskActivity {
	id = 0
	taskId = 0
	userId = 0
	action = ''
	changedFields = {}
	createdAt: Date | null = null
	user: IUser | null = null

	constructor(data: Partial<ITaskActivity> = {}) {
		super()
		this.assignData(data)
		this.createdAt = this.createdAt ? new Date(this.createdAt) : null
		if (typeof data.changedFields === 'string') {
			try {
				this.changedFields = JSON.parse(data.changedFields)
			} catch {
				this.changedFields = {}
			}
		}
		this.user = data.user ? new UserModel(data.user) : null
	}
}
