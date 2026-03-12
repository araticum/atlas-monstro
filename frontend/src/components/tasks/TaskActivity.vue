<template>
	<div class="task-activity details">
		<div class="detail-title">
			<Icon icon="history" />
			Activity
		</div>

		<div
			v-if="loading"
			class="has-text-grey"
		>
			Loading activity…
		</div>
		<div
			v-else-if="activities.length === 0"
			class="has-text-grey"
		>
			No activity yet.
		</div>
		<ul
			v-else
			class="activity-list"
		>
			<li
				v-for="activity in activities"
				:key="activity.id"
				class="activity-item"
			>
				<div class="activity-content">
					<div class="activity-text">
						<strong>{{ displayName(activity) }}</strong>
						{{ formatAction(activity) }}
					</div>
					<div class="activity-date">{{ formatDate(activity.createdAt) }}</div>
				</div>
			</li>
		</ul>
	</div>
</template>

<script setup lang="ts">
import type {ITaskActivity, ITaskActivityChange} from '@/modelTypes/ITaskActivity'
import TaskActivityService from '@/services/taskActivity'

const props = defineProps<{ taskId: number }>()
const activities = ref<ITaskActivity[]>([])
const loading = ref(false)
const service = shallowReactive(new TaskActivityService())

watch(() => props.taskId, async (taskId) => {
	if (!taskId) return
	loading.value = true
	try {
		activities.value = await service.getAll({taskId})
	} finally {
		loading.value = false
	}
}, {immediate: true})

function displayName(activity: ITaskActivity) {
	return activity.user?.name || activity.user?.username || 'Someone'
}

function formatFieldLabel(field: string) {
	return field.replaceAll('_', ' ')
}

function printable(value: unknown) {
	if (value === null || typeof value === 'undefined' || value === '') return 'empty'
	if (Array.isArray(value)) return value.join(', ')
	if (typeof value === 'object') return JSON.stringify(value)
	return String(value)
}

function firstChange(activity: ITaskActivity): [string, ITaskActivityChange] | null {
	const entries = Object.entries(activity.changedFields || {}) as [string, ITaskActivityChange][]
	return entries[0] || null
}

function formatAction(activity: ITaskActivity) {
	const change = firstChange(activity)
	if (activity.action === 'created') return 'created the task'
	if (activity.action === 'custom_field_changed' && change) {
		const [field, values] = change
		return `changed ${formatFieldLabel(field)}: "${printable(values.old)}" → "${printable(values.new)}"`
	}
	if (activity.action === 'commented' && change?.[0] === 'comment') {
		if (change[1].old == null && change[1].new != null) return 'added a comment'
		if (change[1].old != null && change[1].new == null) return 'deleted a comment'
		return 'edited a comment'
	}
	if (change) {
		const [field, values] = change
		return `changed ${formatFieldLabel(field)}: "${printable(values.old)}" → "${printable(values.new)}"`
	}
	return activity.action.replaceAll('_', ' ')
}

function formatDate(value: Date | string | null) {
	if (!value) return ''
	const date = value instanceof Date ? value : new Date(value)
	return new Intl.DateTimeFormat('pt-BR', {
		day: '2-digit',
		month: '2-digit',
		hour: '2-digit',
		minute: '2-digit',
	}).format(date)
}
</script>

<style scoped lang="scss">
.activity-list {
	list-style: none;
	margin: .75rem 0 0;
	padding: 0;
	display: flex;
	flex-direction: column;
	gap: .75rem;
}
.activity-item {
	position: relative;
	padding-left: 1rem;
	border-left: 2px solid var(--grey-200);
}
.activity-content {
	display: flex;
	flex-direction: column;
	gap: .125rem;
}
.activity-text {
	font-size: .95rem;
}
.activity-date {
	font-size: .8rem;
	color: var(--grey-500);
}
</style>
