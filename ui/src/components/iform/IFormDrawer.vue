<template>
  <n-drawer
    :show="iform.drawerVisible"
    placement="right"
    :width="drawerWidth"
    :mask-closable="true"
    :close-on-esc="true"
    @update:show="iform.toggleDrawer"
    class="iform-drawer"
  >
    <n-drawer-content>
      <template #header>
        <div class="iform-drawer__title">
          <n-button v-if="isMobileLayout" size="tiny" quaternary @click="iform.closeDrawer()">
            返回
          </n-button>
          <span>频道嵌入工具</span>
        </div>
      </template>
      <div class="iform-drawer__header">
        <div>
          <p class="iform-drawer__subtitle">可嵌入网页/工具并同步给频道成员</p>
          <div class="iform-drawer__badges">
            <n-tag size="small" type="info">{{ forms.length }} 个控件</n-tag>
            <n-tag size="small" v-if="!iform.canManage" type="warning">只读模式</n-tag>
          </div>
        </div>
        <n-button quaternary size="small" @click="refresh">刷新</n-button>
      </div>

      <n-tabs v-model:value="activeTab" type="line" animated class="iform-workspace-tabs">
        <n-tab-pane name="manage" tab="工具管理">
          <div class="iform-workspace-actions">
            <n-dropdown trigger="click" :options="addMenuOptions" @select="handleAddMenu">
              <n-button type="primary" size="small" :disabled="!iform.canManage">添加工具</n-button>
            </n-dropdown>
            <n-dropdown trigger="click" :options="utilityMenuOptions" @select="handleUtilityMenu">
              <n-button size="small" secondary>更多操作</n-button>
            </n-dropdown>
            <input ref="importInput" type="file" accept="application/json,.json" hidden @change="handleImportFile" />
          </div>

          <n-alert v-if="!iform.canManage" type="info" closable class="iform-permission-alert">
            你当前没有管理权限，但仍可查看和打开频道工具。
          </n-alert>

          <n-spin :show="iform.loading">
            <div v-if="forms.length" class="iform-card-list">
              <article v-for="form in forms" :key="form.id" class="iform-card">
                <div class="iform-card__header">
                  <div class="iform-card__identity">
                    <span class="iform-card__type">{{ form.url ? 'WEB' : 'HTML' }}</span>
                    <div>
                      <strong>{{ form.name || '未命名控件' }}</strong>
                      <p class="iform-card__meta">
                        {{ form.defaultWidth }} × {{ form.defaultHeight }} · {{ form.defaultCollapsed ? '默认折叠' : '默认展开' }}
                      </p>
                    </div>
                  </div>
                  <div class="iform-card__actions">
                    <n-button size="small" secondary @click="openLocal(form, 'top')">打开</n-button>
                    <n-dropdown trigger="click" :options="formMenuOptions(form)" @select="handleFormMenu($event, form)">
                      <n-button size="small" quaternary>更多</n-button>
                    </n-dropdown>
                  </div>
                </div>
                <div class="iform-card__tags">
                  <n-tag v-if="form.templateMissing" size="small" type="error">模板不可用</n-tag>
                  <n-tag v-else-if="form.templateArchived" size="small" type="warning">模板已归档</n-tag>
                  <n-tag v-else-if="form.templateOrigin === 'builtin'" size="small" type="info">内置模板</n-tag>
                  <n-tag v-else-if="form.templateOrigin === 'platform'" size="small" type="info">平台模板</n-tag>
                  <n-tag v-else size="small">独立控件</n-tag>
                  <n-tag v-if="form.worldShared && !form.sharedRef" size="small" type="success">世界共享</n-tag>
                  <n-tag v-if="form.sharedRef" size="small" type="warning">世界引用</n-tag>
                  <n-tag v-if="form.mediaOptions?.autoPlay" size="small">自动播放</n-tag>
                </div>
              </article>
            </div>
            <n-empty v-else description="当前频道暂无嵌入工具">
              <template #extra>
                <n-button v-if="iform.canManage" size="small" type="primary" @click="openTemplateModal">从模板添加</n-button>
              </template>
            </n-empty>
          </n-spin>
        </n-tab-pane>

        <n-tab-pane name="push" tab="推送工具">
          <n-alert v-if="!iform.canBroadcast" type="warning" class="iform-permission-alert">
            你没有向频道推送嵌入工具的权限。
          </n-alert>

          <section class="iform-push-section">
            <div class="iform-section-heading">
              <div>
                <span>01</span>
                <strong>选择工具</strong>
              </div>
              <n-button text size="small" :disabled="!iform.canBroadcast || !forms.length" @click="toggleSelectAllForms">
                {{ allFormsSelected ? '取消全选' : '全选' }}
              </n-button>
            </div>
            <n-checkbox-group :value="iform.selectedFormIds" @update:value="updateSelectedFormIds">
              <div v-if="forms.length" class="iform-push-tool-list">
                <label v-for="form in forms" :key="form.id" class="iform-push-tool" :class="{ 'is-selected': iform.selectedFormIds.includes(form.id) }">
                  <n-checkbox :value="form.id" :disabled="!iform.canBroadcast" />
                  <span>
                    <strong>{{ form.name || '未命名控件' }}</strong>
                    <small>{{ form.defaultWidth }} × {{ form.defaultHeight }}</small>
                  </span>
                </label>
              </div>
              <n-empty v-else description="没有可推送的工具" />
            </n-checkbox-group>
          </section>

          <section class="iform-push-section">
            <div class="iform-section-heading">
              <div><span>02</span><strong>展示方式</strong></div>
            </div>
            <n-radio-group v-model:value="pushModel.placement" size="small" class="iform-placement-options">
              <n-radio-button value="top">上侧面板</n-radio-button>
              <n-radio-button value="right" :disabled="isMobileLayout">右侧面板</n-radio-button>
              <n-radio-button value="floating">浮动窗口</n-radio-button>
            </n-radio-group>
            <p v-if="isMobileLayout" class="iform-field-hint">移动端收到右侧推送时会自动使用上侧面板。</p>

            <div class="iform-push-fields">
              <label class="iform-push-field">
                <span>{{ pushModel.placement === 'floating' ? '初始状态' : '面板状态' }}</span>
                <n-switch v-if="pushModel.placement === 'floating'" v-model:value="pushModel.minimized" size="small">
                  <template #checked>最小化</template>
                  <template #unchecked>展开</template>
                </n-switch>
                <n-switch v-else v-model:value="pushModel.collapsed" size="small">
                  <template #checked>收起</template>
                  <template #unchecked>展开</template>
                </n-switch>
              </label>
              <label class="iform-push-field">
                <span>尺寸</span>
                <n-checkbox v-model:checked="pushModel.useCustomSize">统一设置</n-checkbox>
              </label>
              <div v-if="pushModel.useCustomSize" class="iform-push-size">
                <n-input-number
                  v-if="pushModel.placement !== 'top'"
                  v-model:value="pushModel.width"
                  :min="240"
                  :max="1920"
                  placeholder="宽度"
                />
                <span v-if="pushModel.placement === 'floating'">×</span>
                <n-input-number
                  v-if="pushModel.placement !== 'right'"
                  v-model:value="pushModel.height"
                  :min="160"
                  :max="1440"
                  placeholder="高度"
                />
              </div>
            </div>
          </section>

          <section class="iform-push-section">
            <div class="iform-section-heading">
              <div><span>03</span><strong>推送范围</strong></div>
            </div>
            <n-radio-group v-model:value="pushModel.audience" size="small">
              <n-radio value="all">频道内所有成员</n-radio>
              <n-radio value="members">指定成员</n-radio>
            </n-radio-group>
            <n-select
              v-if="pushModel.audience === 'members'"
              v-model:value="pushModel.targetUserIds"
              multiple
              filterable
              :options="pushTargetOptions"
              placeholder="选择接收成员"
              class="iform-target-select"
            />
          </section>

          <footer class="iform-push-footer">
            <div>
              <span>推送摘要</span>
              <strong>{{ pushSummary }}</strong>
            </div>
            <n-button
              type="primary"
              :loading="pushSubmitting"
              :disabled="!canSubmitPush"
              @click="pushSelected"
            >
              推送到频道
            </n-button>
          </footer>
        </n-tab-pane>
      </n-tabs>

      <n-modal v-model:show="formModalVisible" preset="dialog" :title="editingForm ? '编辑控件' : '新增控件'" :positive-text="editingForm ? '保存' : '创建'" negative-text="取消" @positive-click="handleSubmit" @negative-click="handleCancel">
        <n-form label-placement="left" label-width="72">
          <n-form-item label="名称" required>
            <n-input v-model:value="formModel.name" placeholder="示例：战斗地图" maxlength="64" />
          </n-form-item>
          <n-form-item label="URL">
            <n-input v-model:value="formModel.url" placeholder="https://example.com" :disabled="!!editingForm?.templateRef" />
          </n-form-item>
          <n-form-item label="嵌入代码">
            <n-input type="textarea" v-model:value="formModel.embedCode" placeholder="支持粘贴 HTML / iframe 代码（可含 script）" :rows="3" :disabled="!!editingForm?.templateRef" />
          </n-form-item>
          <n-form-item label="默认尺寸">
            <div class="iform-form__size">
              <n-input-number v-model:value="formModel.defaultWidth" :min="240" :max="1920" placeholder="宽" />
              <span>×</span>
              <n-input-number v-model:value="formModel.defaultHeight" :min="160" :max="1200" placeholder="高" />
            </div>
          </n-form-item>
          <n-form-item label="默认状态">
            <n-switch v-model:value="formModel.defaultCollapsed">
              <template #checked>折叠</template>
              <template #unchecked>展开</template>
            </n-switch>
            <n-switch v-model:value="formModel.defaultFloating">
              <template #checked>弹出</template>
              <template #unchecked>面板</template>
            </n-switch>
            <n-switch v-model:value="formModel.allowPopout">
              <template #checked>允许弹出</template>
              <template #unchecked>禁止弹出</template>
            </n-switch>
          </n-form-item>
          <n-form-item label="媒体优化">
            <n-switch v-model:value="formModel.mediaOptions.autoPlay">
              <template #checked>自动播放</template>
              <template #unchecked>手动播放</template>
            </n-switch>
            <n-switch v-model:value="formModel.mediaOptions.autoUnmute">
              <template #checked>自动解除静音</template>
              <template #unchecked>保持静音</template>
            </n-switch>
            <n-switch v-model:value="formModel.mediaOptions.allowAudio">
              <template #checked>允许音频</template>
              <template #unchecked>禁用音频</template>
            </n-switch>
            <n-switch v-model:value="formModel.mediaOptions.allowVideo">
              <template #checked>允许视频</template>
              <template #unchecked>禁用视频</template>
            </n-switch>
          </n-form-item>
          <n-form-item label="Embed API">
            <n-space vertical size="small">
              <n-switch v-model:value="formModel.bridgePolicy.enabled">
                <template #checked>启用</template>
                <template #unchecked>关闭</template>
              </n-switch>
              <n-input v-model:value="formModel.bridgePolicy.allowedOrigins" placeholder="允许来源，逗号分隔（可选）" :disabled="!formModel.bridgePolicy.enabled" />
              <n-input v-model:value="formModel.bridgePolicy.capabilities" placeholder="能力，逗号分隔（storage.read 等）" :disabled="!formModel.bridgePolicy.enabled" />
            </n-space>
          </n-form-item>
          <n-button v-if="editingForm?.templateRef" size="small" tertiary @click="resetTemplateOverrides">恢复模板默认</n-button>
        </n-form>
      </n-modal>

      <n-modal v-model:show="templateModalVisible" preset="card" title="安装内置工具" style="width: min(620px, 92vw);">
        <n-space vertical>
          <n-input v-model:value="templateSearch" placeholder="搜索模板" clearable @keyup.enter="loadTemplateCatalog(1)" />
          <n-spin :show="templateLoading">
            <n-list bordered>
              <n-list-item v-for="item in templateCatalog" :key="item.ref">
                <n-thing :title="item.name" :description="item.description || ''">
                  <template #header-extra><n-tag size="small" :type="item.origin === 'builtin' ? 'info' : 'success'">{{ item.origin === 'builtin' ? '内置' : '平台' }}</n-tag></template>
                  <template #footer><n-button size="small" type="primary" :disabled="!item.installable" @click="installTemplate(item.ref)">安装</n-button></template>
                </n-thing>
              </n-list-item>
            </n-list>
            <n-pagination
              v-if="templateTotal > templatePageSize"
              v-model:page="templatePage"
              :page-size="templatePageSize"
              :item-count="templateTotal"
              @update:page="loadTemplateCatalog"
            />
          </n-spin>
        </n-space>
      </n-modal>

      <n-modal v-model:show="migrationModalVisible" preset="dialog" title="迁移到其他频道" positive-text="执行" negative-text="取消" @positive-click="handleMigration" @negative-click="() => (migrationModalVisible = false)">
        <n-form label-placement="left" label-width="72">
          <n-form-item label="目标频道" required>
            <n-select v-model:value="migrationTargets" multiple filterable :options="channelOptions" placeholder="选择一个或多个频道" />
          </n-form-item>
          <n-form-item label="模式" required>
            <n-radio-group v-model:value="migrationMode">
              <n-radio value="copy">复制</n-radio>
              <n-radio value="move">迁移</n-radio>
            </n-radio-group>
          </n-form-item>
          <n-form-item label="控件">
            <n-checkbox-group v-model:value="migrationFormIds">
              <n-space vertical>
                <n-checkbox value="@all">全部</n-checkbox>
                <n-checkbox v-for="form in forms" :key="form.id" :value="form.id">{{ form.name || form.id }}</n-checkbox>
              </n-space>
            </n-checkbox-group>
          </n-form-item>
        </n-form>
      </n-modal>
    </n-drawer-content>
  </n-drawer>
</template>

<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue';
import { useWindowSize } from '@vueuse/core';
import { useIFormStore } from '@/stores/iform';
import { useChatStore } from '@/stores/chat';
import { useUtilsStore } from '@/stores/utils';
import { useMessage, useDialog } from 'naive-ui';
import type { ChannelIForm, ChannelIFormPlacement } from '@/types/iform';
import { copyTextWithFallback } from '@/utils/clipboard';
import { generateIFormEmbedLink } from '@/utils/iformEmbedLink';
import { api } from '@/stores/_config';
import type { ChannelIFormTemplateCatalogItem } from '@/types/iform';

const iform = useIFormStore();
const chat = useChatStore();
const utils = useUtilsStore();
iform.bootstrap();

const message = useMessage();
const dialog = useDialog();

const forms = computed(() => [...iform.currentForms]);
const activeTab = ref<'manage' | 'push'>('manage');

const { width: viewportWidth } = useWindowSize();
const drawerWidth = computed(() => {
  if (!viewportWidth.value) return 520;
  return Math.min(560, viewportWidth.value < 640 ? viewportWidth.value : 520);
});
const isMobileLayout = computed(() => viewportWidth.value > 0 && viewportWidth.value < 640);

const formModalVisible = ref(false);
const editingForm = ref<ChannelIForm | null>(null);
const formModel = reactive({
  name: '',
  url: '',
  embedCode: '',
  defaultWidth: 640,
  defaultHeight: 360,
  defaultCollapsed: false,
  defaultFloating: false,
  allowPopout: true,
  mediaOptions: {
    autoPlay: false,
    autoUnmute: false,
    autoExpand: false,
    allowAudio: true,
    allowVideo: true,
  },
  bridgePolicy: {
    enabled: false,
    allowedOrigins: '',
    capabilities: 'context.read,user.read,members.read,world.admins.read,characters.read,permissions.read,storage.read,storage.write,events.subscribe,events.publish,messages.send',
  },
});

const migrationModalVisible = ref(false);
const migrationTargets = ref<string[]>([]);
const migrationMode = ref<'copy' | 'move'>('copy');
const migrationFormIds = ref<string[]>([]);
const templateModalVisible = ref(false);
const templateLoading = ref(false);
const templateSearch = ref('');
const templateCatalog = ref<ChannelIFormTemplateCatalogItem[]>([]);
const templatePage = ref(1);
const templatePageSize = 30;
const templateTotal = ref(0);
const importInput = ref<HTMLInputElement | null>(null);
const pushSubmitting = ref(false);
const pushModel = reactive<{
  placement: ChannelIFormPlacement;
  collapsed: boolean;
  minimized: boolean;
  useCustomSize: boolean;
  width: number;
  height: number;
  audience: 'all' | 'members';
  targetUserIds: string[];
}>({
  placement: 'top',
  collapsed: false,
  minimized: false,
  useCustomSize: false,
  width: 480,
  height: 360,
  audience: 'all',
  targetUserIds: [],
});

const selectedForms = computed(() => forms.value.filter((form) => iform.selectedFormIds.includes(form.id)));
const allFormsSelected = computed(() => forms.value.length > 0 && selectedForms.value.length === forms.value.length);
const pushTargetOptions = computed(() => (chat.curChannelUsers || [])
  .filter((member) => !!member?.id)
  .map((member) => ({
    label: member.nick || member.name || member.id,
    value: member.id,
  })));
const canSubmitPush = computed(() => (
  iform.canBroadcast
  && selectedForms.value.length > 0
  && !pushSubmitting.value
  && (pushModel.audience === 'all' || pushModel.targetUserIds.length > 0)
));
const placementLabels: Record<ChannelIFormPlacement, string> = {
  top: '上侧面板',
  right: '右侧面板',
  floating: '浮动窗口',
};
const pushSummary = computed(() => {
  const toolText = selectedForms.value.length ? `${selectedForms.value.length} 个工具` : '未选择工具';
  const audienceText = pushModel.audience === 'all' ? '全部成员' : `${pushModel.targetUserIds.length} 位成员`;
  const sizeText = pushModel.useCustomSize ? '统一尺寸' : '各自默认尺寸';
  return `${toolText} · ${placementLabels[pushModel.placement]} · ${audienceText} · ${sizeText}`;
});

const addMenuOptions = computed(() => [
  { label: '新建独立工具', key: 'create', disabled: !iform.canManage },
  { label: '从模板安装', key: 'template', disabled: !iform.canManage },
]);
const utilityMenuOptions = computed(() => [
  { label: '迁移或复制', key: 'migrate', disabled: !iform.canManage || !forms.value.length },
  { label: '导入配置', key: 'import', disabled: !iform.canManage },
  { label: '导出配置', key: 'export', disabled: !forms.value.length },
]);

const channelOptions = computed(() => flattenChannels(chat.channelTree || [], chat.curChannel?.id));

function flattenChannels(tree: any[], excludeId?: string, depth = 0): Array<{ label: string; value: string }> {
  const result: Array<{ label: string; value: string }> = [];
  tree.forEach((node) => {
    if (!node?.id || node.id === excludeId) {
      return;
    }
    const indent = depth ? `${'· '.repeat(depth)}` : '';
    result.push({ label: `${indent}${node.name || node.id}`, value: node.id });
    if (node.children?.length) {
      result.push(...flattenChannels(node.children, excludeId, depth + 1));
    }
  });
  return result;
}

watch(
  () => iform.visibleChannelId,
  () => {
    pushModel.audience = 'all';
    pushModel.targetUserIds = [];
  },
);

const handleAddMenu = (key: string) => {
  if (key === 'create') {
    openFormModal();
  } else if (key === 'template') {
    void openTemplateModal();
  }
};

const handleUtilityMenu = (key: string) => {
  if (key === 'migrate') {
    migrationModalVisible.value = true;
  } else if (key === 'import') {
    openImport();
  } else if (key === 'export') {
    exportForms();
  }
};

const formMenuOptions = (form: ChannelIForm) => [
  { label: '在右侧打开', key: 'open-right', disabled: isMobileLayout.value },
  { label: '作为浮窗打开', key: 'open-floating' },
  { label: '复制嵌入链接', key: 'copy-link' },
  { type: 'divider', key: 'display-divider' },
  { label: '配置并推送', key: 'prepare-push', disabled: !iform.canBroadcast },
  { label: '编辑配置', key: 'edit', disabled: !iform.canManage || !canEditForm(form) },
  {
    label: form.worldShared ? '取消世界共享' : '共享到世界',
    key: 'world-share',
    disabled: !iform.canManageWorldShared || !!form.sharedRef || iform.isReadonlyForm(form),
  },
  { type: 'divider', key: 'danger-divider' },
  { label: '删除工具', key: 'delete', disabled: !iform.canManage || !canDeleteForm(form) },
];

const handleFormMenu = (key: string, form: ChannelIForm) => {
  if (key === 'open-right') {
    openLocal(form, 'right');
  } else if (key === 'open-floating') {
    openLocal(form, 'floating');
  } else if (key === 'copy-link') {
    void copyEmbedLink(form);
  } else if (key === 'prepare-push') {
    iform.setSelected([form.id]);
    activeTab.value = 'push';
  } else if (key === 'edit') {
    openFormModal(form);
  } else if (key === 'world-share') {
    void toggleWorldShare(form);
  } else if (key === 'delete') {
    confirmDelete(form);
  }
};

const openLocal = (form: ChannelIForm, placement: ChannelIFormPlacement) => {
  if (placement === 'floating') {
    openFloating(form.id);
    return;
  }
  iform.openPanel(form.id, {
    placement,
    width: form.defaultWidth,
    height: form.defaultHeight,
    collapsed: !!form.defaultCollapsed,
  });
};

const updateSelectedFormIds = (values: Array<string | number>) => {
  iform.setSelected(values.map(String));
};

const toggleSelectAllForms = () => {
  iform.setSelected(allFormsSelected.value ? [] : forms.value.map((form) => form.id));
};

const resetFormModel = () => {
  editingForm.value = null;
  Object.assign(formModel, {
    name: '',
    url: '',
    embedCode: '',
    defaultWidth: 640,
    defaultHeight: 360,
    defaultCollapsed: false,
    defaultFloating: false,
    allowPopout: true,
    mediaOptions: {
      autoPlay: false,
      autoUnmute: false,
      autoExpand: false,
      allowAudio: true,
      allowVideo: true,
    },
    bridgePolicy: {
      enabled: false,
      allowedOrigins: '',
      capabilities: 'context.read,user.read,members.read,world.admins.read,characters.read,permissions.read,storage.read,storage.write,events.subscribe,events.publish,messages.send',
    },
  });
};

const openFormModal = (form?: ChannelIForm) => {
  if (!iform.canManage) {
    return;
  }
  if (form && !canEditForm(form)) {
    return;
  }
  if (form) {
    editingForm.value = form;
    Object.assign(formModel, {
      name: form.name,
      url: form.url || '',
      embedCode: form.embedCode || '',
      defaultWidth: form.defaultWidth || 640,
      defaultHeight: form.defaultHeight || 360,
      defaultCollapsed: !!form.defaultCollapsed,
      defaultFloating: !!form.defaultFloating,
      allowPopout: form.allowPopout !== false,
      mediaOptions: {
        autoPlay: !!form.mediaOptions?.autoPlay,
        autoUnmute: !!form.mediaOptions?.autoUnmute,
        autoExpand: !!form.mediaOptions?.autoExpand,
        allowAudio: form.mediaOptions?.allowAudio !== false,
        allowVideo: form.mediaOptions?.allowVideo !== false,
      },
      bridgePolicy: {
        enabled: !!form.bridgePolicy?.enabled,
        allowedOrigins: (form.bridgePolicy?.allowedOrigins || []).join(','),
        capabilities: (form.bridgePolicy?.capabilities || []).join(','),
      },
    });
  } else {
    resetFormModel();
  }
  formModalVisible.value = true;
};

const handleSubmit = async () => {
  if (!formModel.name.trim()) {
    message.warning('名称不能为空');
    return false;
  }
  if (!editingForm.value?.templateRef && !formModel.url.trim() && !formModel.embedCode.trim()) {
    message.warning('请至少填写 URL 或嵌入代码');
    return false;
  }
  try {
    if (editingForm.value) {
      const payload: Record<string, unknown> = {};
      if (editingForm.value.templateRef) {
        const original = editingForm.value;
        const mediaOptions = { ...formModel.mediaOptions };
        const originalMediaOptions = {
          autoPlay: !!original.mediaOptions?.autoPlay,
          autoUnmute: !!original.mediaOptions?.autoUnmute,
          autoExpand: !!original.mediaOptions?.autoExpand,
          allowAudio: original.mediaOptions?.allowAudio !== false,
          allowVideo: original.mediaOptions?.allowVideo !== false,
        };
        const bridgePolicy = normalizeBridgePolicyForm();
        const originalBridgePolicy = {
          enabled: !!original.bridgePolicy?.enabled,
          allowedOrigins: original.bridgePolicy?.allowedOrigins || [],
          capabilities: original.bridgePolicy?.capabilities || [],
        };
        if (formModel.name.trim() !== (original.name || '')) payload.name = formModel.name.trim();
        if (formModel.defaultWidth !== original.defaultWidth) payload.defaultWidth = formModel.defaultWidth;
        if (formModel.defaultHeight !== original.defaultHeight) payload.defaultHeight = formModel.defaultHeight;
        if (formModel.defaultCollapsed !== !!original.defaultCollapsed) payload.defaultCollapsed = formModel.defaultCollapsed;
        if (formModel.defaultFloating !== !!original.defaultFloating) payload.defaultFloating = formModel.defaultFloating;
        if (formModel.allowPopout !== original.allowPopout) payload.allowPopout = formModel.allowPopout;
        if (JSON.stringify(mediaOptions) !== JSON.stringify(originalMediaOptions)) payload.mediaOptions = mediaOptions;
        if (JSON.stringify(bridgePolicy) !== JSON.stringify(originalBridgePolicy)) payload.bridgePolicy = bridgePolicy;
      } else {
        payload.name = formModel.name.trim();
        payload.defaultWidth = formModel.defaultWidth;
        payload.defaultHeight = formModel.defaultHeight;
        payload.defaultCollapsed = formModel.defaultCollapsed;
        payload.defaultFloating = formModel.defaultFloating;
        payload.allowPopout = formModel.allowPopout;
        payload.mediaOptions = formModel.mediaOptions;
        payload.bridgePolicy = normalizeBridgePolicyForm();
        payload.url = formModel.url.trim();
        payload.embedCode = formModel.embedCode.trim();
      }
      await iform.updateForm(editingForm.value.id, payload);
      message.success('控件已更新');
    } else {
      await iform.createForm({
        name: formModel.name.trim(),
        url: formModel.url.trim(),
        embedCode: formModel.embedCode.trim(),
        defaultWidth: formModel.defaultWidth,
        defaultHeight: formModel.defaultHeight,
        defaultCollapsed: formModel.defaultCollapsed,
        defaultFloating: formModel.defaultFloating,
        allowPopout: formModel.allowPopout,
        mediaOptions: formModel.mediaOptions,
        bridgePolicy: normalizeBridgePolicyForm(),
      });
      message.success('控件已创建');
    }
    formModalVisible.value = false;
    resetFormModel();
    return true;
  } catch (error: any) {
    message.error(error?.response?.data?.message || error?.message || '保存失败');
    return false;
  }
};

const normalizeBridgePolicyForm = () => ({
  enabled: !!formModel.bridgePolicy.enabled,
  allowedOrigins: formModel.bridgePolicy.allowedOrigins.split(',').map((item) => item.trim()).filter(Boolean),
  capabilities: formModel.bridgePolicy.capabilities.split(',').map((item) => item.trim()).filter(Boolean),
});

const handleCancel = () => {
  resetFormModel();
  return true;
};

const confirmDelete = (form: ChannelIForm) => {
  if (!iform.canManage || !canDeleteForm(form)) {
    return;
  }
  dialog.warning({
    title: '删除控件',
    content: `确认删除「${form.name || form.id}」？该操作不可撤销。`,
    positiveText: '删除',
    negativeText: '取消',
    async onPositiveClick() {
      try {
        await iform.deleteForm(form.id);
        message.success('已删除');
      } catch (error: any) {
        message.error(error?.response?.data?.message || '删除失败');
      }
    },
  });
};

const openFloating = (formId: string) => {
  if (!formId) {
    return;
  }
  const windowId = iform.createWindowId(formId);
  iform.openFloating(formId, { windowId });
};

const resolveEmbedLinkBase = () => {
  const domain = utils.config?.domain?.trim() || '';
  if (!domain) {
    return undefined;
  }
  const webUrl = utils.config?.webUrl?.trim() || '';
  let base = domain;
  if (!/^(https?:)?\/\//i.test(base)) {
    base = `${window.location.protocol}//${base}`;
  }
  if (webUrl) {
    base = `${base}${webUrl.startsWith('/') ? '' : '/'}${webUrl}`;
  }
  return base;
};

const copyEmbedLink = async (form: ChannelIForm) => {
  const worldId = chat.currentWorldId;
  const channelId = chat.curChannel?.id;
  if (!worldId || !channelId || !form?.id) {
    message.warning('无法生成嵌入链接');
    return;
  }
  const link = generateIFormEmbedLink(
    {
      worldId,
      channelId: form.sourceChannelId || form.channelId,
      formId: form.id,
      width: form.defaultWidth,
      height: form.defaultHeight,
    },
    { base: resolveEmbedLinkBase() },
  );
  const copied = await copyTextWithFallback(link);
  if (copied) {
    message.success('嵌入链接已复制');
  } else {
    message.error('复制失败');
  }
};

const pushSelected = async () => {
  if (!canSubmitPush.value) {
    return;
  }
  const states = selectedForms.value.map((form) => ({
      formId: form.id,
      placement: pushModel.placement,
      width: pushModel.useCustomSize ? pushModel.width : form.defaultWidth,
      height: pushModel.useCustomSize ? pushModel.height : form.defaultHeight,
      collapsed: pushModel.placement === 'floating' ? false : pushModel.collapsed,
      floating: pushModel.placement === 'floating',
      minimized: pushModel.placement === 'floating' && pushModel.minimized,
    }));
  pushSubmitting.value = true;
  try {
    await iform.pushStates(states, {
      force: true,
      targetUserIds: pushModel.audience === 'members' ? pushModel.targetUserIds : undefined,
    });
    message.success(`已推送 ${states.length} 个频道工具`);
  } catch (error: any) {
    message.error(error?.response?.data?.message || '推送失败');
  } finally {
    pushSubmitting.value = false;
  }
};

const canEditForm = (form: ChannelIForm) => !iform.isReadonlyForm(form);

const canDeleteForm = (form: ChannelIForm) => !form.sharedRef && !iform.isReadonlyForm(form);

const toggleWorldShare = async (form: ChannelIForm) => {
  if (!iform.canManageWorldShared) {
    return;
  }
  const enabled = !form.worldShared;
  try {
    await iform.toggleWorldShare([form.id], enabled);
    message.success(enabled ? '已共享到世界' : '已取消世界共享');
  } catch (error: any) {
    message.error(error?.response?.data?.message || error?.message || '世界共享切换失败');
  }
};

const refresh = async () => {
  if (!iform.currentChannelId) {
    return;
  }
  await iform.ensureForms(iform.currentChannelId, true);
  message.success('已刷新控件列表');
};

const loadTemplateCatalog = async (page = templatePage.value) => {
  templatePage.value = page;
  templateLoading.value = true;
  try {
    const { data } = await api.get<{ items: ChannelIFormTemplateCatalogItem[]; total?: number }>('api/v1/channel-embed-tools/catalog', {
      params: { search: templateSearch.value.trim(), page: templatePage.value, pageSize: templatePageSize },
    });
    templateCatalog.value = data?.items || [];
    templateTotal.value = data?.total || 0;
  } catch (error: any) {
    message.error(error?.response?.data?.message || '读取模板目录失败');
  } finally {
    templateLoading.value = false;
  }
};

const openTemplateModal = async () => {
  templateModalVisible.value = true;
  templatePage.value = 1;
  await loadTemplateCatalog(1);
};

const installTemplate = async (templateRef: string) => {
  try {
    await iform.createForm({ name: '', templateRef });
    await iform.ensureForms(iform.visibleChannelId || '', true);
    message.success('模板已安装');
    templateModalVisible.value = false;
  } catch (error: any) {
    message.error(error?.response?.data?.message || error?.message || '安装失败');
  }
};

const resetTemplateOverrides = async () => {
  if (!editingForm.value?.templateRef) return;
  try {
    await iform.updateForm(editingForm.value.id, { templateOverrides: {} });
    message.success('已恢复模板默认');
    formModalVisible.value = false;
    resetFormModel();
  } catch (error: any) {
    message.error(error?.response?.data?.message || '恢复失败');
  }
};

const exportForms = () => {
  const items = forms.value.map((form) => form.templateRef
    ? { mode: 'reference', templateRef: form.templateRef, overrides: form.templateOverrides || {}, local: { orderIndex: form.orderIndex } }
    : { mode: 'standalone', config: {
      name: form.name, url: form.url || '', embedCode: form.embedCode || '', defaultWidth: form.defaultWidth,
      defaultHeight: form.defaultHeight, defaultCollapsed: form.defaultCollapsed, defaultFloating: form.defaultFloating,
      allowPopout: form.allowPopout, mediaOptions: form.mediaOptions, bridgePolicy: form.bridgePolicy,
    } });
  const blob = new Blob([JSON.stringify({ schemaVersion: 1, type: 'sealchat-channel-iforms', items }, null, 2)], { type: 'application/json' });
  const url = URL.createObjectURL(blob);
  const anchor = document.createElement('a');
  anchor.href = url;
  anchor.download = 'sealchat-channel-iforms.json';
  anchor.click();
  URL.revokeObjectURL(url);
};

const openImport = () => importInput.value?.click();

const handleImportFile = async (event: Event) => {
  const input = event.target as HTMLInputElement;
  const file = input.files?.[0];
  input.value = '';
  if (!file) return;
  try {
    const bundle = JSON.parse(await file.text());
    if (bundle?.schemaVersion !== 1 || bundle?.type !== 'sealchat-channel-iforms' || !Array.isArray(bundle.items)) {
      throw new Error('导入文件格式无效');
    }
    let success = 0;
    const failures: string[] = [];
    for (const [index, item] of bundle.items.entries()) {
      try {
        if (item?.mode === 'reference' && typeof item.templateRef === 'string') {
          await iform.createForm({ templateRef: item.templateRef, templateOverrides: item.overrides || {}, orderIndex: item.local?.orderIndex || 0, name: '' });
        } else if (item?.mode === 'standalone' && item.config) {
          await iform.createForm(item.config);
        } else {
          throw new Error('项目模式无效');
        }
        success += 1;
      } catch (error: any) {
        failures.push(`${index + 1}: ${error?.response?.data?.message || error?.message || '失败'}`);
      }
    }
    await iform.ensureForms(iform.visibleChannelId || '', true);
    message.info(`导入完成：成功 ${success}，失败 ${failures.length}${failures.length ? `（${failures.join('；')}）` : ''}`);
  } catch (error: any) {
    message.error(error?.message || '导入失败');
  }
};

const handleMigration = async () => {
  try {
    const targets = migrationTargets.value.slice();
    const selected = migrationFormIds.value.includes('@all') ? [] : migrationFormIds.value;
    await iform.migrateForms(targets, selected, migrationMode.value);
    message.success('迁移任务已提交');
    migrationModalVisible.value = false;
    migrationTargets.value = [];
    migrationFormIds.value = [];
  } catch (error: any) {
    message.error(error?.response?.data?.message || '迁移失败');
    return false;
  }
  return true;
};
</script>

<style scoped>
.iform-drawer :deep(.n-drawer-body) {
  background: var(--sc-bg-elevated, #0f172a);
  color: var(--sc-text-primary, #e2e8f0);
}

.iform-drawer__title {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.iform-drawer__header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 1rem;
}

.iform-drawer__subtitle {
  margin: 0;
  font-size: 0.9rem;
  color: var(--sc-text-secondary, rgba(226, 232, 240, 0.8));
}

.iform-drawer__badges {
  display: flex;
  gap: 0.35rem;
  margin-top: 0.35rem;
}

.iform-workspace-tabs {
  min-height: 0;
}

.iform-workspace-actions {
  display: flex;
  align-items: center;
  gap: 0.55rem;
  margin-bottom: 0.9rem;
}

.iform-permission-alert {
  margin-bottom: 0.9rem;
}

.iform-card-list {
  display: grid;
  gap: 0.65rem;
}

.iform-card {
  position: relative;
  overflow: hidden;
  padding: 0.9rem;
  border: 1px solid var(--iform-card-border, var(--sc-border-mute, rgba(148, 163, 184, 0.24)));
  border-radius: 14px;
  background: var(--iform-card-bg, var(--sc-bg-surface, rgba(248, 250, 252, 0.72)));
  color: var(--iform-card-text, var(--sc-text-primary, #0f172a));
  transition: border-color 160ms ease, transform 160ms ease;
}

.iform-card::before {
  content: '';
  position: absolute;
  top: 0;
  bottom: 0;
  left: 0;
  width: 3px;
  background: var(--sc-primary-color, #3b82f6);
  opacity: 0.55;
}

.iform-card:hover {
  border-color: rgba(var(--sc-primary-rgb, 59, 130, 246), 0.34);
  transform: translateY(-1px);
}

.iform-card strong {
  color: inherit;
}

.iform-card__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.8rem;
}

.iform-card__identity {
  display: flex;
  align-items: center;
  gap: 0.7rem;
  min-width: 0;
}

.iform-card__identity > div {
  min-width: 0;
}

.iform-card__identity strong {
  display: block;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.iform-card__type {
  display: grid;
  width: 2.45rem;
  height: 2.45rem;
  flex: 0 0 auto;
  place-items: center;
  border: 1px solid rgba(var(--sc-primary-rgb, 59, 130, 246), 0.28);
  border-radius: 10px;
  background: rgba(var(--sc-primary-rgb, 59, 130, 246), 0.09);
  color: var(--sc-primary-color, #2563eb);
  font-size: 0.58rem;
  font-weight: 800;
  letter-spacing: 0.08em;
}

.iform-card__meta {
  margin: 0.15rem 0 0;
  color: var(--sc-text-secondary, rgba(100, 116, 139, 0.9));
  font-size: 0.76rem;
}

.iform-card__tags {
  display: flex;
  flex-wrap: wrap;
  gap: 0.35rem;
  margin-top: 0.65rem;
  padding-left: 3.15rem;
}

.iform-card__actions {
  display: flex;
  flex: 0 0 auto;
  gap: 0.2rem;
}

.iform-push-section {
  margin-bottom: 0.8rem;
  padding: 0.9rem;
  border: 1px solid var(--sc-border-mute, rgba(148, 163, 184, 0.22));
  border-radius: 14px;
  background: var(--sc-bg-surface, rgba(248, 250, 252, 0.58));
}

.iform-section-heading {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 0.75rem;
}

.iform-section-heading > div {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.iform-section-heading span {
  color: var(--sc-primary-color, #2563eb);
  font-size: 0.65rem;
  font-weight: 800;
  letter-spacing: 0.08em;
}

.iform-push-tool-list {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.45rem;
}

.iform-push-tool {
  display: flex;
  align-items: center;
  gap: 0.55rem;
  min-width: 0;
  padding: 0.62rem;
  border: 1px solid var(--sc-border-mute, rgba(148, 163, 184, 0.2));
  border-radius: 10px;
  background: var(--sc-bg-elevated, rgba(255, 255, 255, 0.72));
  cursor: pointer;
  transition: border-color 150ms ease, background 150ms ease;
}

.iform-push-tool.is-selected {
  border-color: rgba(var(--sc-primary-rgb, 59, 130, 246), 0.48);
  background: rgba(var(--sc-primary-rgb, 59, 130, 246), 0.08);
}

.iform-push-tool > span {
  display: flex;
  min-width: 0;
  flex-direction: column;
}

.iform-push-tool strong {
  overflow: hidden;
  color: var(--sc-text-primary, #0f172a);
  font-size: 0.82rem;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.iform-push-tool small,
.iform-field-hint {
  color: var(--sc-text-secondary, #64748b);
  font-size: 0.7rem;
}

.iform-placement-options {
  display: flex;
  width: 100%;
}

.iform-placement-options :deep(.n-radio-button) {
  flex: 1;
  text-align: center;
}

.iform-field-hint {
  margin: 0.45rem 0 0;
}

.iform-push-fields {
  display: grid;
  gap: 0.7rem;
  margin-top: 0.85rem;
  padding-top: 0.8rem;
  border-top: 1px dashed var(--sc-border-mute, rgba(148, 163, 184, 0.26));
}

.iform-push-field {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  color: var(--sc-text-secondary, #64748b);
  font-size: 0.78rem;
}

.iform-push-size {
  display: flex;
  align-items: center;
  gap: 0.45rem;
}

.iform-push-size :deep(.n-input-number) {
  flex: 1;
}

.iform-target-select {
  margin-top: 0.7rem;
}

.iform-push-footer {
  position: sticky;
  z-index: 2;
  bottom: -1rem;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.8rem;
  margin: 0.25rem -0.25rem -1rem;
  padding: 0.85rem 0.25rem 1rem;
  border-top: 1px solid var(--sc-border-mute, rgba(148, 163, 184, 0.22));
  background: var(--sc-bg-elevated, #ffffff);
}

.iform-push-footer > div {
  display: flex;
  min-width: 0;
  flex-direction: column;
}

.iform-push-footer span {
  color: var(--sc-text-secondary, #64748b);
  font-size: 0.66rem;
}

.iform-push-footer strong {
  overflow: hidden;
  color: var(--sc-text-primary, #0f172a);
  font-size: 0.75rem;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.iform-form__size {
  display: flex;
  align-items: center;
  gap: 0.35rem;
}

@media (max-width: 639px) {
  .iform-push-tool-list {
    grid-template-columns: 1fr;
  }

  .iform-card__header,
  .iform-push-footer {
    align-items: stretch;
    flex-direction: column;
  }

  .iform-card__actions {
    justify-content: flex-end;
  }

  .iform-card__tags {
    padding-left: 0;
  }
}
</style>
