<template>
	<div class="task-templates details content">
		<h3>
			<span class="icon is-grey">
				<Icon icon="clone" />
			</span>
			{{ $t('task.templates.title') }}
		</h3>

		<p
			v-if="!loading && templates.length === 0"
			class="has-text-grey"
		>
			{{ $t('task.templates.empty') }}
		</p>

		<div
			v-for="template in templates"
			:key="template.id"
			class="template-card"
		>
			<div>
				<strong>{{ template.title }}</strong>
				<p
					v-if="template.description"
					class="template-description"
				>
					{{ template.description }}
				</p>
			</div>
			<XButton
				variant="secondary"
				icon="copy"
				:loading="loadingTemplateId === template.id"
				@click="useTemplate(template.id)"
			>
				{{ $t('task.templates.use') }}
			</XButton>
		</div>
	</div>
</template>

<script setup lang="ts">
import {onMounted, ref, watch} from 'vue'
import {useI18n} from 'vue-i18n'

import type {ITask} from '@/modelTypes/ITask'
import TaskTemplateService from '@/services/taskTemplate'
import {success} from '@/message'

const props = defineProps<{
	projectId: ITask['projectId']
	currentTaskId?: ITask['id']
}>()

const emit = defineEmits<{
	(e: 'taskCreated', task: ITask): void
}>()

const {t} = useI18n({useScope: 'global'})
const service = new TaskTemplateService()
const templates = ref<ITask[]>([])
const loading = ref(false)
const loadingTemplateId = ref<number | null>(null)

async function loadTemplates() {
	if (!props.projectId) {
		return
	}

	loading.value = true
	try {
		const loaded = await service.getAll({
			projectId: props.projectId,
		} as ITask)
		templates.value = loaded.filter(template => template.id !== props.currentTaskId)
	} finally {
		loading.value = false
	}
}

async function useTemplate(templateId: number) {
	loadingTemplateId.value = templateId
	try {
		const task = await service.duplicate(props.projectId, templateId)
		success({message: t('task.templates.createdSuccess')})
		emit('taskCreated', task)
		await loadTemplates()
	} finally {
		loadingTemplateId.value = null
	}
}

watch(() => [props.projectId, props.currentTaskId], loadTemplates)
onMounted(loadTemplates)
</script>

<style scoped lang="scss">
.template-card {
	display: flex;
	justify-content: space-between;
	gap: 1rem;
	align-items: flex-start;
	padding: 0.75rem 0;
	border-block-end: 1px solid var(--grey-200);
}

.template-description {
	margin: 0.25rem 0 0;
	color: var(--grey-600);
	white-space: pre-wrap;
}
</style>
