<template>
	<CreateEdit
		v-if="uploadBackgroundEnabled || unsplashBackgroundEnabled"
		:title="$t('project.background.title')"
		:loading="backgroundService.loading"
		class="project-background-setting"
		:wide="true"
	>
		<div
			v-if="uploadBackgroundEnabled"
			class="mbe-4"
		>
			<input
				ref="backgroundUploadInput"
				accept="image/*"
				class="is-hidden"
				type="file"
				@change="uploadBackground"
			>
			<XButton
				:loading="backgroundUploadService.loading"
				variant="primary"
				@click="backgroundUploadInput?.click()"
			>
				{{ $t('project.background.upload') }}
			</XButton>
		</div>
		<template v-if="unsplashBackgroundEnabled">
			<input
				v-model="backgroundSearchTerm"
				:class="{'is-loading': backgroundService.loading}"
				class="input is-expanded"
				:placeholder="$t('project.background.searchPlaceholder')"
				type="text"
				@keyup="debounceNewBackgroundSearch()"
			>

			<p class="unsplash-credit">
				<BaseButton
					class="unsplash-credit__link"
					href="https://unsplash.com"
				>
					{{ $t('project.background.poweredByUnsplash') }}
				</BaseButton>
			</p>

			<ul class="image-search__result-list">
				<li
					v-for="im in backgroundSearchResult"
					:key="im.id"
					class="image-search__result-item"
					:style="{'background-image': `url(${backgroundBlurHashes[im.id]})` }"
				>
					<CustomTransition name="fade">
						<BaseButton
							v-if="backgroundThumbs[im.id]"
							class="image-search__image-button"
							@click="setBackground(im.id)"
						>
							<img
								class="image-search__image"
								:src="backgroundThumbs[im.id]"
								alt=""
							>
						</BaseButton>
					</CustomTransition>

					<BaseButton
						:href="`https://unsplash.com/@${im.info.author}?utm_source=vikunja&utm_medium=referral`"
						class="image-search__info"
					>
						{{ im.info.authorName }}
					</BaseButton>
				</li>
			</ul>
			<XButton
				v-if="backgroundSearchResult.length > 0"
				:disabled="backgroundService.loading"
				class="is-load-more-button mbs-4"
				:shadow="false"
				variant="secondary"
				@click="searchBackgrounds(currentPage + 1)"
			>
				{{ backgroundService.loading ? $t('misc.loading') : $t('project.background.loadMore') }}
			</XButton>
		</template>

		<template #footer>
			<XButton
				v-if="hasBackground"
				:shadow="false"
				variant="tertiary"
				danger
				@click.prevent.stop="removeBackground"
			>
				{{ $t('project.background.remove') }}
			</XButton>
			<XButton
				variant="secondary"
				@click.prevent.stop="$router.back()"
			>
				{{ $t('misc.close') }}
			</XButton>
		</template>
	</CreateEdit>
</template>


<script setup lang="ts">
import {ref, computed, shallowReactive} from 'vue'
import {useI18n} from 'vue-i18n'
import {useRoute, useRouter} from 'vue-router'
import {useDebounceFn} from '@vueuse/core'

import BaseButton from '@/components/base/BaseButton.vue'
import CustomTransition from '@/components/misc/CustomTransition.vue'

import {useBaseStore} from '@/stores/base'
import {useProjectStore} from '@/stores/projects'
import {useConfigStore} from '@/stores/config'

import BackgroundUnsplashService from '@/services/backgroundUnsplash'
import BackgroundUploadService from '@/services/backgroundUpload'
import ProjectService from '@/services/project'

import {getBlobFromBlurHash} from '@/helpers/getBlobFromBlurHash'
import {useTitle} from '@/composables/useTitle'

import CreateEdit from '@/components/misc/CreateEdit.vue'
import {success} from '@/message'

import type {IBackgroundImage} from '@/modelTypes/IBackgroundImage'
import type {IProject} from '@/modelTypes/IProject'
import type {IFile} from '@/modelTypes/IFile'

defineOptions({name: 'ProjectSettingBackground'})

const SEARCH_DEBOUNCE = 300

const {t} = useI18n({useScope: 'global'})
const baseStore = useBaseStore()
const route = useRoute()
const router = useRouter()

useTitle(() => t('project.background.title'))

const backgroundService = shallowReactive(new BackgroundUnsplashService())
const backgroundSearchTerm = ref('')
const backgroundSearchResult = ref<IBackgroundImage[]>([])
const backgroundThumbs = ref<Record<number, string>>({})
const backgroundBlurHashes = ref<Record<number, string>>({})
const currentPage = ref(1)

const debounceNewBackgroundSearch = useDebounceFn(newBackgroundSearch, SEARCH_DEBOUNCE)

const backgroundUploadService = shallowReactive(new BackgroundUploadService())
const projectService = shallowReactive(new ProjectService())
const projectStore = useProjectStore()
const configStore = useConfigStore()

const unsplashBackgroundEnabled = computed(() => configStore.enabledBackgroundProviders.includes('unsplash'))
const uploadBackgroundEnabled = computed(() => configStore.enabledBackgroundProviders.includes('upload'))
const currentProject = computed(() => baseStore.currentProject)
const hasBackground = computed(() => !!currentProject.value?.backgroundInformation)
const projectId = computed(() => Number(route.params.projectId))

newBackgroundSearch()

function newBackgroundSearch() {
	if (!unsplashBackgroundEnabled.value) {
		return
	}
	backgroundSearchResult.value = []
	backgroundThumbs.value = {}
	backgroundBlurHashes.value = {}
	searchBackgrounds()
}

async function searchBackgrounds(page = 1) {
	currentPage.value = page
	const result = await backgroundService.getAll({} as unknown as IBackgroundImage, {s: backgroundSearchTerm.value, p: page}) as IBackgroundImage[]
	backgroundSearchResult.value = backgroundSearchResult.value.concat(result)
	result.forEach((background) => {
		getBlobFromBlurHash(background.blurHash)
			.then((blob) => {
				if (!blob) {
					return
				}
				backgroundBlurHashes.value[background.id] = window.URL.createObjectURL(blob)
			})

		backgroundService.thumb(background).then(thumb => {
			backgroundThumbs.value[background.id] = thumb
		})
	})
}

async function setBackground(backgroundId: IBackgroundImage['id']) {
	if (backgroundService.loading) {
		return
	}

	const project = await backgroundService.update({
		id: backgroundId,
		projectId: projectId.value,
	} as unknown as IBackgroundImage) as unknown as IProject
	await baseStore.handleSetCurrentProject({project, forceUpdate: true})
	projectStore.setProject(project)
	success({message: t('project.background.success')})
}

const backgroundUploadInput = ref<HTMLInputElement | null>(null)
async function uploadBackground() {
	const file = backgroundUploadInput.value?.files?.[0]
	if (!file) {
		return
	}

	const project = await backgroundUploadService.create(projectId.value, file as unknown as IFile) as unknown as IProject
	await baseStore.handleSetCurrentProject({project, forceUpdate: true})
	projectStore.setProject(project)
	success({message: t('project.background.success')})
}

async function removeBackground() {
	if (!currentProject.value) {
		return
	}

	const project = await projectService.removeBackground(currentProject.value as IProject)
	await baseStore.handleSetCurrentProject({project, forceUpdate: true})
	projectStore.setProject(project)
	success({message: t('project.background.removeSuccess')})
	router.back()
}
</script>

<style lang="scss" scoped>
.unsplash-credit {
	text-align: end;
	font-size: .8rem;
}

.unsplash-credit__link {
	color: var(--grey-800);
}

.image-search__result-list {
	--items-per-row: 1;
	margin: 1rem 0 0;
	display: grid;
	gap: 1rem;
	grid-template-columns: repeat(var(--items-per-row), 1fr);

	@media screen and (min-width: $mobile) {
		--items-per-row: 2;
	}
	@media screen and (min-width: $tablet) {
		--items-per-row: 4;
	}
	@media screen and (min-width: $tablet) {
		--items-per-row: 5;
	}
}

.image-search__result-item {
	margin-block-start: 0; // FIXME: removes padding from .content
	aspect-ratio: 16 / 10;
	background-size: cover;
	background-position: center;
	display: flex;
	position: relative;
}

.image-search__image-button {
	inline-size: 100%;
}

.image-search__image {
	inline-size: 100%;
	block-size: 100%;
	object-fit: cover;
}

.image-search__info {
	position: absolute;
	inset-block-end: 0;
	inline-size: 100%;
	padding: .25rem 0;
	opacity: 0;
	text-align: center;
	background: rgba(0, 0, 0, 0.5);
	font-size: .75rem;
	font-weight: bold;
	color: $white;
	transition: opacity $transition;
}
.image-search__result-item:hover .image-search__info {
		opacity: 1;
}

.is-load-more-button {
	margin: 1rem auto 0 !important;
	display: block;
	inline-size: 200px;
}
</style>
