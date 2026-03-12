import AbstractService from './abstractService'
import TaskModel from '@/models/task'
import type { ITask } from '@/modelTypes/ITask'
import { AuthenticatedHTTPFactory } from '@/helpers/fetcher'

export default class TaskTemplateService extends AbstractService<ITask> {
	constructor() {
		super({
			getAll: '/projects/{projectId}/task-templates',
		})
	}

	modelFactory(data: Partial<ITask>) {
		return new TaskModel(data)
	}

	async duplicate(
		projectId: ITask['projectId'],
		templateId: ITask['id'],
	): Promise<ITask> {
		const cancel = this.setLoading()
		try {
			const response = await AuthenticatedHTTPFactory().post(
				`/projects/${projectId}/task-templates/${templateId}/duplicate`,
				{},
			)
			return this.modelFactory(response.data)
		} finally {
			cancel()
		}
	}

	async setIsTemplate(
		taskId: ITask['id'],
		isTemplate: boolean,
	): Promise<ITask> {
		const cancel = this.setLoading()
		try {
			const response = await AuthenticatedHTTPFactory().put(
				`/tasks/${taskId}/template`,
				{ is_template: isTemplate },
			)
			return this.modelFactory(response.data)
		} finally {
			cancel()
		}
	}
}
