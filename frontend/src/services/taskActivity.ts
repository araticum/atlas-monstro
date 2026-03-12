import AbstractService from './abstractService'
import TaskActivityModel from '@/models/taskActivity'
import type {ITaskActivity} from '@/modelTypes/ITaskActivity'

export default class TaskActivityService extends AbstractService<ITaskActivity> {
	constructor() {
		super({
			getAll: '/tasks/{taskId}/activities',
		})
	}

	modelFactory(data) {
		return new TaskActivityModel(data)
	}
}
