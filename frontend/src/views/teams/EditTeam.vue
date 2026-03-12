<template>
	<div
		class="loader-container is-max-width-desktop"
		:class="{ 'is-loading': teamService.loading }"
	>
		<Card
			v-if="team && userIsAdmin && !teamOidcId"
			class="is-fullwidth"
			:title="title"
		>
			<form @submit.prevent="save()">
				<FormField
					id="teamtext"
					v-model="team.name"
					v-focus
					:label="$t('team.attributes.name')"
					:disabled="teamMemberService.loading"
					:loading="teamMemberService.loading"
					:placeholder="$t('team.attributes.namePlaceholder')"
					type="text"
					:error="showErrorTeamnameRequired && team.name === '' ? $t('team.attributes.nameRequired') : null"
				/>
				<FormField
					v-if="configStore.publicTeamsEnabled"
					:label="$t('team.attributes.isPublic')"
				>
					<FancyCheckbox
						v-model="team.isPublic"
						:disabled="teamMemberService.loading || undefined"
						:class="{ 'disabled': teamService.loading }"
					>
						{{ $t('team.attributes.isPublicDescription') }}
					</FancyCheckbox>
				</FormField>
				<FormField :label="$t('team.attributes.description')">
					<Editor
						id="teamdescription"
						v-model="team.description"
						:class="{ disabled: teamService.loading }"
						:disabled="teamService.loading"
						:placeholder="$t('team.attributes.descriptionPlaceholder')"
					/>
				</FormField>

				<div class="field has-addons mbs-4">
					<div class="control is-fullwidth">
						<XButton
							:loading="teamService.loading"
							class="is-fullwidth"
							type="submit"
						>
							{{ $t('misc.save') }}
						</XButton>
					</div>
					<div class="control">
						<XButton
							:loading="teamService.loading"
							danger
							icon="trash-alt"
							@click="showDeleteModal = true"
						/>
					</div>
				</div>
			</form>
		</Card>

		<Card
			v-if="team"
			class="is-fullwidth has-overflow"
			:title="$t('team.edit.members')"
			:padding="false"
		>
			<form
				v-if="userIsAdmin && !teamOidcId"
				class="p-4"
				@submit.prevent="addUser"
			>
				<div class="field has-addons">
					<div class="control is-expanded">
						<Multiselect
							v-model="newMember"
							:loading="userService.loading"
							:placeholder="$t('team.edit.search')"
							:search-results="foundUsers"
							label="username"
							@search="findUser"
						>
							<template #searchResult="{option: user}">
								<User
									v-if="isUserOption(user)"
									:avatar-size="24"
									:user="user"
									class="m-0"
								/>
							</template>
						</Multiselect>
					</div>
					<div class="control">
						<XButton
							icon="plus"
							@click="addUser"
						>
							{{ $t('team.edit.addUser') }}
						</XButton>
					</div>
				</div>
				<p
					v-if="showMustSelectUserError"
					class="help is-danger"
				>
					{{ $t('team.edit.mustSelectUser') }}
				</p>
			</form>
			<div class="has-horizontal-overflow">
				<table class="table has-actions is-striped is-hoverable is-fullwidth">
					<tbody>
						<tr
							v-for="m in team.members"
							:key="m.id"
						>
							<td>
								<User
									:avatar-size="24"
									:user="m"
									class="m-0"
								/>
							</td>
							<td>
								<template v-if="m.id === userInfo?.id">
									<b class="is-success">You</b>
								</template>
							</td>
							<td class="type">
								<template v-if="m.admin">
									<span class="icon is-small">
										<Icon icon="lock" />
									</span>
									{{ $t('team.attributes.admin') }}
								</template>
								<template v-else>
									<span class="icon is-small">
										<Icon icon="user" />
									</span>
									{{ $t('team.attributes.member') }}
								</template>
							</td>
							<td
								v-if="userIsAdmin"
								class="actions"
							>
								<XButton
									v-if="m.id !== userInfo?.id"
									:loading="teamMemberService.loading"
									class="mie-2"
									@click="() => toggleUserType(m)"
								>
									{{ m.admin ? $t('team.edit.makeMember') : $t('team.edit.makeAdmin') }}
								</XButton>
								<XButton
									v-if="m.id !== userInfo?.id"
									:loading="teamMemberService.loading"
									danger
									icon="trash-alt"
									@click="() => {memberToDelete = m; showUserDeleteModal = true}"
								/>
							</td>
						</tr>
					</tbody>
				</table>
			</div>
		</Card>

		<XButton
			v-if="team && !team.externalId"
			class="is-fullwidth is-danger"
			@click="showLeaveModal = true"
		>
			{{ $t('team.edit.leave.title') }}
		</XButton>

		<Modal
			v-if="showLeaveModal"
			@close="showLeaveModal = false"
			@submit="leave()"
		>
			<template #header>
				<span>{{ $t('team.edit.leave.title') }}</span>
			</template>

			<template #text>
				<p>
					{{ $t('team.edit.leave.text1') }}<br>
					{{ $t('team.edit.leave.text2') }}
				</p>
			</template>
		</Modal>

		<Modal
			:enabled="showDeleteModal"
			@close="showDeleteModal = false"
			@submit="deleteTeam()"
		>
			<template #header>
				<span>{{ $t('team.edit.delete.header') }}</span>
			</template>

			<template #text>
				<p>
					{{ $t('team.edit.delete.text1') }}<br>
					{{ $t('team.edit.delete.text2') }}
				</p>
			</template>
		</Modal>

		<Modal
			:enabled="showUserDeleteModal"
			@close="showUserDeleteModal = false"
			@submit="deleteMember()"
		>
			<template #header>
				<span>{{ $t('team.edit.deleteUser.header') }}</span>
			</template>

			<template #text>
				<p>
					{{ $t('team.edit.deleteUser.text1') }}<br>
					{{ $t('team.edit.deleteUser.text2') }}
				</p>
			</template>
		</Modal>
	</div>
</template>

<script lang="ts" setup>
import {computed, ref} from 'vue'
import {useI18n} from 'vue-i18n'
import {useRoute, useRouter} from 'vue-router'

import Editor from '@/components/input/AsyncEditor'
import FancyCheckbox from '@/components/input/FancyCheckbox.vue'
import FormField from '@/components/input/FormField.vue'
import Multiselect from '@/components/input/Multiselect.vue'
import User from '@/components/misc/User.vue'

import TeamMemberModel from '@/models/teamMember'
import TeamService from '@/services/team'
import TeamMemberService from '@/services/teamMember'
import UserService from '@/services/user'

import {PERMISSIONS as Permissions} from '@/constants/permissions'

import {useTitle} from '@/composables/useTitle'
import {success} from '@/message'
import {useAuthStore} from '@/stores/auth'
import {useConfigStore} from '@/stores/config'

import type {ITeam} from '@/modelTypes/ITeam'
import type {IUser} from '@/modelTypes/IUser'
import type {ITeamMember} from '@/modelTypes/ITeamMember'

type SearchUser = IUser & Record<string, unknown>

const authStore = useAuthStore()
const configStore = useConfigStore()
const route = useRoute()
const router = useRouter()
const {t} = useI18n({useScope: 'global'})

const team = ref<ITeam | null>(null)
const teamOidcId = computed(() => Boolean((team.value as (ITeam & {oidcId?: string | null}) | null)?.oidcId))
const userIsAdmin = computed(() => (
	team.value !== null &&
	team.value.maxPermission !== null &&
	team.value.maxPermission > Permissions.READ
))
const userInfo = computed(() => authStore.info)

const teamService = ref(new TeamService())
const teamMemberService = ref(new TeamMemberService())
const userService = ref(new UserService())

const teamId = computed(() => Number(route.params.id))
const memberToDelete = ref<ITeamMember | null>(null)
const newMember = ref<SearchUser | null>(null)
const foundUsers = ref<SearchUser[]>([])

const showDeleteModal = ref(false)
const showUserDeleteModal = ref(false)
const showLeaveModal = ref(false)
const showErrorTeamnameRequired = ref(false)
const showMustSelectUserError = ref(false)

const title = ref('')

loadTeam()

async function loadTeam() {
	team.value = await teamService.value.get({id: teamId.value} as ITeam)
	title.value = t('team.edit.title', {team: team.value?.name})
	useTitle(() => title.value)
}

async function save() {
	if (!team.value) return
	if (team.value.name === '') {
		showErrorTeamnameRequired.value = true
		return
	}
	showErrorTeamnameRequired.value = false

	team.value = await teamService.value.update(team.value)
	success({message: t('team.edit.success')})
}

async function deleteTeam() {
	if (!team.value) return
	await teamService.value.delete(team.value)
	success({message: t('team.edit.delete.success')})
	router.push({name: 'teams.index'})
}

async function deleteMember() {
	if (!memberToDelete.value) return
	try {
		await teamMemberService.value.delete(new TeamMemberModel({
			teamId: teamId.value,
			username: memberToDelete.value.username,
		}))
		success({message: t('team.edit.deleteUser.success')})
		await loadTeam()
	} finally {
		showUserDeleteModal.value = false
		memberToDelete.value = null
	}
}

async function addUser() {
	showMustSelectUserError.value = false
	if(!newMember.value) {
		showMustSelectUserError.value = true
		return
	}
	await teamMemberService.value.create(new TeamMemberModel({
		teamId: teamId.value,
		username: newMember.value.username,
	}))
	newMember.value = null
	await loadTeam()
	success({message: t('team.edit.userAddedSuccess')})
}

async function toggleUserType(member: ITeamMember) {
	if (!team.value) return
	member.admin = !member.admin
	member.teamId = teamId.value
	const r = await teamMemberService.value.update(member)
	for (const tm of team.value.members) {
		if (tm.id === member.id) {
			tm.admin = r.admin
			break
		}
	}
	success({
		message: member.admin ? t('team.edit.madeAdmin') : t('team.edit.madeMember'),
	})
}

function isUserOption(option: string | SearchUser): option is SearchUser {
	return typeof option !== 'string'
}

async function findUser(query: string) {
	if (query === '') {
		foundUsers.value = []
		return
	}

	const users = await userService.value.getAll({} as IUser, {s: query})
	foundUsers.value = users.filter((u: IUser) => u.id !== userInfo.value?.id) as SearchUser[]
}

async function leave() {
	if (!userInfo.value) return
	try {
		await teamMemberService.value.delete(new TeamMemberModel({
			teamId: teamId.value,
			username: userInfo.value.username,
		}))
		success({message: t('team.edit.leave.success')})
		await router.push({name: 'home'})
	} finally {
		showLeaveModal.value = false
	}
}
</script>

<style lang="scss" scoped>
.card.is-fullwidth {
	margin-block-end: 1rem;

	.content {
		padding: 0;
	}
}
</style>
