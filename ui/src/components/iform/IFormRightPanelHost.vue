<template>
  <teleport to="body">
    <aside
      v-if="panels.length && activePanel"
      class="iform-right-panel"
      :class="{ 'is-collapsed': activePanel.collapsed, 'is-resizing': !!resizing }"
      :style="panelStyle"
      aria-label="右侧频道嵌入面板"
    >
    <button
      v-if="activePanel.collapsed"
      type="button"
      class="iform-right-panel__rail"
      :title="`展开 ${resolveForm(activePanel.formId)?.name || '频道嵌入工具'}`"
      @click="toggleCollapse(activePanel.windowId)"
    >
      <n-icon :component="ChevronBackOutline" size="17" />
      <span>{{ resolveForm(activePanel.formId)?.name || '嵌入工具' }}</span>
      <em v-if="panels.length > 1">{{ panels.length }}</em>
    </button>

    <div
      v-else
      class="iform-right-panel__card"
      :class="{ 'has-tabs': panels.length > 1 }"
    >
      <header class="iform-right-panel__header">
        <div class="iform-right-panel__heading">
          <span class="iform-right-panel__eyebrow">CHANNEL TOOL</span>
          <div class="iform-right-panel__title">
            <strong>{{ resolveForm(activePanel.formId)?.name || '未命名嵌入' }}</strong>
            <n-tag v-if="activePanel.fromPush" size="small" type="info">同步</n-tag>
          </div>
        </div>
        <div class="iform-right-panel__actions">
          <n-tooltip trigger="hover">
            <template #trigger>
              <n-button quaternary size="tiny" @click="moveToTop(activePanel.windowId)">
                <template #icon><n-icon :component="ArrowUpOutline" /></template>
              </n-button>
            </template>
            移到上侧
          </n-tooltip>
          <n-tooltip trigger="hover">
            <template #trigger>
              <n-button quaternary size="tiny" @click="openFloating(activePanel)">
                <template #icon><n-icon :component="OpenOutline" /></template>
              </n-button>
            </template>
            弹出窗口
          </n-tooltip>
          <n-tooltip trigger="hover">
            <template #trigger>
              <n-button quaternary size="tiny" :disabled="!iform.canBroadcast" @click="pushCurrent(activePanel)">
                <template #icon><n-icon :component="ShareOutline" /></template>
              </n-button>
            </template>
            推送当前布局
          </n-tooltip>
          <n-button quaternary size="tiny" title="收起到右侧" @click="toggleCollapse(activePanel.windowId)">
            <template #icon><n-icon :component="ChevronForwardOutline" /></template>
          </n-button>
          <n-button quaternary size="tiny" title="关闭" @click="closePanel(activePanel.windowId)">
            <template #icon><n-icon :component="CloseOutline" /></template>
          </n-button>
        </div>
      </header>

      <nav v-if="panels.length > 1" class="iform-right-panel__tabs" aria-label="已打开的频道工具">
        <button
          v-for="panel in panels"
          :key="panel.windowId"
          type="button"
          :class="{ 'is-active': panel.windowId === activeWindowId }"
          @click="activatePanel(panel.windowId)"
        >
          <span>{{ resolveForm(panel.formId)?.name || '未命名工具' }}</span>
          <i v-if="panel.fromPush" aria-label="同步推送" />
        </button>
      </nav>

      <div class="iform-right-panel__body">
        <section
          v-for="panel in panels"
          :key="panel.windowId"
          v-show="panel.windowId === activeWindowId"
          class="iform-right-panel__surface"
        >
          <div v-if="panel.autoPlayHint || panel.autoUnmuteHint" class="iform-right-panel__banner">
            <n-icon size="14" :component="VolumeHighOutline" />
            <span>页面含媒体时，可能需要点击后播放或解除静音。</span>
          </div>
          <IFormEmbedPortal
            v-if="!panel.collapsed"
            :window-id="panel.windowId"
            :form-id="panel.formId"
            surface="right"
          />
        </section>
      </div>
      <div class="iform-right-panel__resize" @mousedown.prevent="startResizing(activePanel, $event)">
        <span />
      </div>
    </div>
    </aside>
  </teleport>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import { useEventListener, useWindowSize } from '@vueuse/core';
import { useMessage } from 'naive-ui';
import {
  ArrowUpOutline,
  ChevronBackOutline,
  ChevronForwardOutline,
  CloseOutline,
  OpenOutline,
  ShareOutline,
  VolumeHighOutline,
} from '@vicons/ionicons5';
import { useIFormStore } from '@/stores/iform';
import type { ChannelIForm } from '@/types/iform';
import IFormEmbedPortal from './IFormEmbedPortal.vue';

const iform = useIFormStore();
iform.bootstrap();

const message = useMessage();
const panels = computed(() => iform.currentRightPanels);
const formMap = computed(() => new Map(iform.currentForms.map((form) => [form.id, form])));
const activeWindowId = ref('');
const resizing = ref<{ windowId: string; startWidth: number; startX: number } | null>(null);
const { width: viewportWidth } = useWindowSize();

const activePanel = computed(() => (
  panels.value.find((panel) => panel.windowId === activeWindowId.value)
  || panels.value[panels.value.length - 1]
  || null
));
const panelStyle = computed(() => ({
  width: activePanel.value?.collapsed ? '44px' : `${activePanel.value?.width || 420}px`,
}));

const resolveForm = (formId: string): ChannelIForm | undefined => formMap.value.get(formId);

watch(
  () => panels.value.map((panel) => panel.windowId),
  (ids, previousIds = []) => {
    const added = ids.find((id) => !previousIds.includes(id));
    if (added) {
      activeWindowId.value = added;
      return;
    }
    if (!ids.includes(activeWindowId.value)) {
      activeWindowId.value = ids[ids.length - 1] || '';
    }
  },
  { immediate: true },
);

watch(viewportWidth, (width) => {
  if (width > 0 && width < 768) {
    [...panels.value].forEach((panel) => iform.movePanel(panel.windowId, 'top'));
  }
});

const activatePanel = (windowId: string) => {
  activeWindowId.value = windowId;
  const panel = panels.value.find((item) => item.windowId === windowId);
  if (panel?.collapsed) {
    iform.togglePanelCollapse(windowId);
  }
};

const toggleCollapse = (windowId: string) => iform.togglePanelCollapse(windowId);
const closePanel = (windowId: string) => {
  resizing.value = null;
  iform.closePanel(windowId);
};
const moveToTop = (windowId: string) => {
  resizing.value = null;
  iform.movePanel(windowId, 'top');
};
const openFloating = (panel: NonNullable<typeof activePanel.value>) => {
  const form = resolveForm(panel.formId);
  resizing.value = null;
  iform.openFloating(panel.formId, {
    windowId: panel.windowId,
    width: form?.defaultWidth || panel.width,
    height: form?.defaultHeight || panel.height,
    fromPush: false,
  });
};
const pushCurrent = async (panel: NonNullable<typeof activePanel.value>) => {
  if (!iform.canBroadcast) return;
  try {
    await iform.pushStates([{
      formId: panel.formId,
      placement: 'right',
      floating: false,
      collapsed: panel.collapsed,
      width: panel.width,
      height: panel.height,
    }], { force: true });
    message.success('已按右侧布局推送到频道');
  } catch (error: any) {
    message.error(error?.response?.data?.message || '推送失败');
  }
};

const startResizing = (panel: NonNullable<typeof activePanel.value>, event: MouseEvent) => {
  resizing.value = { windowId: panel.windowId, startWidth: panel.width, startX: event.clientX };
};

useEventListener(window, 'mousemove', (event: MouseEvent) => {
  if (!resizing.value) return;
  event.preventDefault();
  iform.resizePanelWidth(
    resizing.value.windowId,
    resizing.value.startWidth + resizing.value.startX - event.clientX,
  );
});
useEventListener(window, 'mouseup', () => { resizing.value = null; });
useEventListener(window, 'blur', () => { resizing.value = null; });
</script>

<style scoped>
.iform-right-panel {
  position: fixed;
  z-index: 31;
  top: 4.35rem;
  right: 1rem;
  bottom: 1rem;
  min-width: 300px;
  max-width: min(720px, 62vw);
  transition: width 180ms ease, transform 180ms ease;
}

.iform-right-panel.is-collapsed {
  min-width: 44px;
}

.iform-right-panel.is-resizing {
  transition: none;
  user-select: none;
}

.iform-right-panel__card {
  position: relative;
  display: grid;
  grid-template-rows: auto minmax(0, 1fr);
  width: 100%;
  height: 100%;
  overflow: hidden;
  border: 1px solid var(--sc-border-strong, rgba(15, 23, 42, 0.14));
  border-radius: 18px;
  background: var(--sc-bg-elevated, #ffffff);
  color: var(--sc-text-primary, #0f172a);
  box-shadow: -18px 22px 54px rgba(15, 23, 42, 0.18);
}

.iform-right-panel__card.has-tabs {
  grid-template-rows: auto auto minmax(0, 1fr);
}

.iform-right-panel__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  padding: 0.75rem 0.8rem 0.7rem 1rem;
  border-bottom: 1px solid var(--sc-border-mute, rgba(15, 23, 42, 0.08));
}

.iform-right-panel__heading {
  min-width: 0;
}

.iform-right-panel__eyebrow {
  display: block;
  margin-bottom: 0.12rem;
  color: var(--sc-text-secondary, #64748b);
  font-size: 0.58rem;
  font-weight: 700;
  letter-spacing: 0.16em;
}

.iform-right-panel__title {
  display: flex;
  align-items: center;
  gap: 0.45rem;
  min-width: 0;
}

.iform-right-panel__title strong {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.iform-right-panel__actions {
  display: flex;
  flex: 0 0 auto;
  gap: 0.15rem;
}

.iform-right-panel__tabs {
  display: flex;
  gap: 0.35rem;
  padding: 0.5rem 0.75rem;
  overflow-x: auto;
  border-bottom: 1px solid var(--sc-border-mute, rgba(15, 23, 42, 0.08));
  background: var(--sc-bg-surface, rgba(248, 250, 252, 0.86));
}

.iform-right-panel__tabs button {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  min-width: 0;
  max-width: 11rem;
  padding: 0.32rem 0.62rem;
  border: 1px solid transparent;
  border-radius: 999px;
  background: transparent;
  color: var(--sc-text-secondary, #64748b);
  font-size: 0.74rem;
  cursor: pointer;
}

.iform-right-panel__tabs button span {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.iform-right-panel__tabs button i {
  width: 0.38rem;
  height: 0.38rem;
  flex: 0 0 auto;
  border-radius: 50%;
  background: #0ea5e9;
  box-shadow: 0 0 0 3px rgba(14, 165, 233, 0.13);
}

.iform-right-panel__tabs button.is-active {
  border-color: rgba(var(--sc-primary-rgb, 59, 130, 246), 0.28);
  background: rgba(var(--sc-primary-rgb, 59, 130, 246), 0.1);
  color: var(--sc-text-primary, #0f172a);
}

.iform-right-panel__body,
.iform-right-panel__surface {
  position: relative;
  min-height: 0;
  height: 100%;
}

.iform-right-panel__surface {
  padding: 0.75rem;
}

.iform-right-panel__banner {
  position: absolute;
  z-index: 2;
  top: 1rem;
  left: 1rem;
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  max-width: calc(100% - 2rem);
  padding: 0.25rem 0.55rem;
  border-radius: 999px;
  background: rgba(14, 165, 233, 0.14);
  color: var(--sc-primary-color, #0369a1);
  font-size: 0.72rem;
}

.iform-right-panel__resize {
  position: absolute;
  z-index: 3;
  top: 0;
  bottom: 0;
  left: -0.35rem;
  width: 0.7rem;
  cursor: ew-resize;
}

.iform-right-panel__resize span {
  position: absolute;
  top: calc(50% - 2rem);
  left: 0.2rem;
  width: 0.22rem;
  height: 4rem;
  border-radius: 999px;
  background: rgba(var(--sc-primary-rgb, 59, 130, 246), 0.38);
  opacity: 0;
  transition: opacity 150ms ease;
}

.iform-right-panel__resize:hover span,
.iform-right-panel.is-resizing .iform-right-panel__resize span {
  opacity: 1;
}

.iform-right-panel__rail {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.6rem;
  width: 44px;
  min-height: 10rem;
  padding: 0.75rem 0.45rem;
  border: 1px solid var(--sc-border-strong, rgba(15, 23, 42, 0.14));
  border-radius: 14px 0 0 14px;
  background: var(--sc-bg-elevated, #ffffff);
  color: var(--sc-text-primary, #0f172a);
  box-shadow: -10px 14px 34px rgba(15, 23, 42, 0.16);
  cursor: pointer;
}

.iform-right-panel__rail span {
  overflow: hidden;
  writing-mode: vertical-rl;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 0.75rem;
  font-weight: 600;
  letter-spacing: 0.08em;
}

.iform-right-panel__rail em {
  display: grid;
  width: 1.2rem;
  height: 1.2rem;
  place-items: center;
  border-radius: 50%;
  background: var(--sc-primary-color, #3b82f6);
  color: #fff;
  font-size: 0.65rem;
  font-style: normal;
}

@media (max-width: 767px) {
  .iform-right-panel {
    display: none;
  }
}
</style>
