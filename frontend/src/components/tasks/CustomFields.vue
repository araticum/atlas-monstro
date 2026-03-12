<template>
	<div
		v-if="schemas.length > 0"
		class="details custom-fields"
	>
		<h3>
			<span class="icon is-grey">
				<Icon icon="list-alt" />
			</span>
			Custom Fields
		</h3>

		<div
			v-for="schema in schemas"
			:key="schema.id"
			class="field"
		>
			<label class="label">{{ schema.name }}</label>
			<p
				v-if="schema.description"
				class="help"
			>
				{{ schema.description }}
			</p>

			<div class="control">
				<input
					v-if="schema.fieldType === 'text'"
					v-model="localValues[schema.id]"
					class="input"
					:type="schema.fieldType === 'number' ? 'number' : 'text'"
					:disabled="disabled"
					@blur="save(schema)"
				>
				<input
					v-else-if="schema.fieldType === 'number'"
					v-model="localValues[schema.id]"
					class="input"
					type="number"
					:disabled="disabled"
					@blur="save(schema)"
				>
				<input
					v-else-if="schema.fieldType === 'date'"
					v-model="localValues[schema.id]"
					class="input"
					type="date"
					:disabled="disabled"
					@change="save(schema)"
				>
				<select
					v-else-if="schema.fieldType === 'select'"
					v-model="localValues[schema.id]"
					class="input"
					:disabled="disabled"
					@change="save(schema)"
				>
					<option value="">
						—
					</option>
					<option
						v-for="option in getOptions(schema)"
						:key="option.value"
						:value="option.value"
					>
						{{ option.label || option.value }}
					</option>
				</select>
				<label
					v-else-if="schema.fieldType === 'checkbox'"
					class="checkbox"
				>
					<input
						v-model="checkboxValues[schema.id]"
						type="checkbox"
						:disabled="disabled"
						@change="save(schema)"
					>
					{{ schema.required ? 'Required' : 'Optional' }}
				</label>
				<textarea
					v-else-if="schema.fieldType === 'textarea'"
					v-model="localValues[schema.id]"
					class="textarea"
					:disabled="disabled"
					@blur="save(schema)"
				/>
			</div>
		</div>
	</div>
</template>

<script setup lang="ts">
import {ref, watch} from 'vue'
import {AuthenticatedHTTPFactory} from '@/helpers/fetcher'
import {success} from '@/message'

type Schema = {
	id: number
	name: string
	fieldType: string
	description?: string
	options?: string
	required?: boolean
	defaultValue?: string
}

type CustomFieldValue = {
	field_id?: number
	fieldId?: number
	value: unknown
}

type CustomFieldsResponse = {
	schemas?: Array<Schema & {
		field_type?: string
		default_value?: string
	}>
	values?: CustomFieldValue[]
}

const props = defineProps<{
	taskId: number
	disabled?: boolean
}>()

const http = AuthenticatedHTTPFactory()
const schemas = ref<Schema[]>([])
const localValues = ref<Record<number, string>>({})
const checkboxValues = ref<Record<number, boolean>>({})

watch(() => props.taskId, load, {immediate: true})

async function load() {
	if (!props.taskId) {
		return
	}

	const {data} = await http.get<CustomFieldsResponse>(`/tasks/${props.taskId}/custom-fields`)
	schemas.value = (data.schemas || []).map(schema => ({
		...schema,
		fieldType: schema.field_type ?? schema.fieldType,
		defaultValue: schema.default_value ?? schema.defaultValue,
	}))

	const valuesByFieldId = Object.fromEntries((data.values || []).map(value => [
		value.field_id ?? value.fieldId,
		value.value,
	]))

	const nextValues: Record<number, string> = {}
	const nextChecks: Record<number, boolean> = {}
	for (const schema of schemas.value) {
		const raw = valuesByFieldId[schema.id] ?? schema.defaultValue ?? ''
		nextValues[schema.id] = String(raw ?? '')
		nextChecks[schema.id] = raw === true || raw === 'true'
	}
	localValues.value = nextValues
	checkboxValues.value = nextChecks
}

function getOptions(schema: Schema): Array<{value: string, label?: string}> {
	try {
		const parsed = JSON.parse(schema.options || '[]')
		return Array.isArray(parsed) ? parsed : (parsed.options || [])
	} catch {
		return []
	}
}

async function save(schema: Schema) {
	if (props.disabled) {
		return
	}

	const value = schema.fieldType === 'checkbox'
		? checkboxValues.value[schema.id]
		: localValues.value[schema.id]

	if (value === '' || value === null || typeof value === 'undefined') {
		await http.delete(`/tasks/${props.taskId}/custom-fields/${schema.id}`)
		return
	}

	await http.put(`/tasks/${props.taskId}/custom-fields/${schema.id}`, {value})
	success({message: 'Custom field saved.'})
}
</script>

<style scoped lang="scss">
.custom-fields {
	margin-block: 1.5rem;
}
</style>
