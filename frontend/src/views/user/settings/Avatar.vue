<template>
	<Card :title="$t('user.settings.avatar.title')">
		<Message v-if="avatarProvider === 'ldap'">
			{{ $t('user.settings.avatar.ldap') }}
		</Message>

		<Message v-else-if="avatarProvider === 'openid'">
			{{ $t('user.settings.avatar.openid', {provider: authProvider}) }}
		</Message>

		<template v-else>
			<div class="control mbe-4">
				<label
					v-for="(label, providerId) in AVATAR_PROVIDERS"
					:key="providerId"
					class="radio"
				>
					<input
						v-model="avatarProvider"
						name="avatarProvider"
						type="radio"
						:value="providerId"
					>
					{{ label }}
				</label>
			</div>

			<template v-if="avatarProvider === 'upload'">
				<input
					ref="avatarUploadInput"
					accept="image/*"
					class="is-hidden"
					type="file"
					@change="cropAvatar"
				>

				<XButton
					v-if="!isCropAvatar"
					:loading="avatarService.loading || loading"
					@click="avatarUploadInput?.click()"
				>
					{{ $t('user.settings.avatar.uploadAvatar') }}
				</XButton>
				<template v-else>
					<Cropper
						ref="cropper"
						:src="avatarToCrop"
						:stencil-props="{aspectRatio: 1}"
						class="mbe-4 cropper"
						@ready="() => loading = false"
					/>
					<XButton
						v-cy="'uploadAvatar'"
						:loading="avatarService.loading || loading"
						@click="uploadAvatar"
					>
						{{ $t('user.settings.avatar.uploadAvatar') }}
					</XButton>
				</template>
			</template>

			<div
				v-else
				class="mbs-2"
			>
				<XButton
					:loading="avatarService.loading || loading"
					class="is-fullwidth"
					@click="updateAvatarStatus()"
				>
					{{ $t('misc.save') }}
				</XButton>
			</div>
		</template>
	</Card>
</template>


<script setup lang="ts">
import {computed, ref, shallowReactive} from 'vue'
import {useI18n} from 'vue-i18n'
import {Cropper} from 'vue-advanced-cropper'
import 'vue-advanced-cropper/dist/style.css'

import AvatarService from '@/services/avatar'
import AvatarModel from '@/models/avatar'
import {useTitle} from '@/composables/useTitle'
import {success} from '@/message'
import {useAuthStore} from '@/stores/auth'
import Message from '@/components/misc/Message.vue'

import type {AvatarProvider, IAvatar} from '@/modelTypes/IAvatar'


defineOptions({name: 'UserSettingsAvatar'})

const {t} = useI18n({useScope: 'global'})
const authStore = useAuthStore()
const authProvider = computed(() => ((authStore.info as ({authProvider?: string} & typeof authStore.info) | null)?.authProvider) ?? '')

const AVATAR_PROVIDERS = computed<Record<AvatarProvider, string>>(() => ({
	default: t('misc.default'),
	initials: t('user.settings.avatar.initials'),
	gravatar: t('user.settings.avatar.gravatar'),
	marble: t('user.settings.avatar.marble'),
	upload: t('user.settings.avatar.upload'),
}))

useTitle(() => `${t('user.settings.avatar.title')} - ${t('user.settings.title')}`)

const avatarService = shallowReactive(new AvatarService())
const loading = ref(false)
const avatarProvider = ref<AvatarProvider | 'ldap' | 'openid'>('default')

async function avatarStatus() {
	const currentAvatar = await avatarService.get(new AvatarModel({}) as IAvatar)
	avatarProvider.value = currentAvatar.avatarProvider
}

avatarStatus()

async function updateAvatarStatus() {
	await avatarService.update(new AvatarModel({avatarProvider: avatarProvider.value as AvatarProvider}))
	success({message: t('user.settings.avatar.statusUpdateSuccess')})
	authStore.reloadAvatar()
}

type CropperResult = { canvas?: HTMLCanvasElement | null }
type CropperLike = { getResult: () => CropperResult }

const cropper = ref<CropperLike | null>(null)
const isCropAvatar = ref(false)

async function uploadAvatar() {
	loading.value = true
	const {canvas} = cropper.value?.getResult() ?? {}

	if (!canvas) {
		loading.value = false
		return
	}

	try {
		const blob = await new Promise<Blob | null>((resolve) => canvas.toBlob((canvasBlob: Blob | null) => resolve(canvasBlob)))
		if (!blob) {
			return
		}
		await avatarService.create(blob)
		success({message: t('user.settings.avatar.setSuccess')})
		authStore.reloadAvatar()
	} finally {
		loading.value = false
		isCropAvatar.value = false
	}
}

const avatarToCrop = ref<string>('')
const avatarUploadInput = ref<HTMLInputElement | null>(null)

function cropAvatar() {
	const avatar = avatarUploadInput.value?.files
	if (!avatar || avatar.length === 0) {
		return
	}

	loading.value = true
	const reader = new FileReader()
	reader.onload = (e: ProgressEvent<FileReader>) => {
		avatarToCrop.value = typeof e.target?.result === 'string' ? e.target.result : ''
		isCropAvatar.value = avatarToCrop.value !== ''
	}
	reader.onerror = () => {
		loading.value = false
	}
	reader.readAsDataURL(avatar[0])
}
</script>

<style lang="scss">
.cropper {
	block-size: 80vh;
	background: transparent;
}

.vue-advanced-cropper__background {
	background: var(--white);
}
</style>
