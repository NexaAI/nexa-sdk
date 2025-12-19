package ai.nexa.agent.ui.theme


import androidx.compose.material3.ColorScheme
import androidx.compose.runtime.Composable
import androidx.compose.ui.graphics.Color


// Text
val ColorScheme.textPrimary: Color
    @Composable get() = if (LocalAppDarkTheme.current) TextPrimaryDark else TextPrimaryLight

val ColorScheme.textSecondary: Color
    @Composable get() = if (LocalAppDarkTheme.current) TextSecondaryDark else TextSecondaryLight


// Splash
val ColorScheme.SplashBackground: Color
    @Composable get() = if (LocalAppDarkTheme.current) SplashBackgroundDark else SplashBackgroundLight

// Drawer
val ColorScheme.drawerBackground: Color
    @Composable get() = if (LocalAppDarkTheme.current) DrawerBackgroundDark else DrawerBackgroundLight

val ColorScheme.outlinedButtonBorder: Color
    @Composable get() = if (LocalAppDarkTheme.current) OutlinedButtonBorderDark else OutlinedButtonBorderLight

val ColorScheme.drawerIcon: Color
    @Composable get() = if (LocalAppDarkTheme.current) DrawerIconDark else DrawerIconLight

val ColorScheme.dividerColor: Color
    @Composable get() = if (LocalAppDarkTheme.current) DividerColorDark else DividerColorLight

val ColorScheme.sessionActive: Color
    @Composable get() = if (LocalAppDarkTheme.current) SessionActiveDark else SessionActiveLight

val ColorScheme.dialogBackground: Color
    @Composable get() = if (LocalAppDarkTheme.current) DialogBackgroundDark else DialogBackgroundLight

val ColorScheme.dialogContainerColor: Color
    @Composable get() = if (LocalAppDarkTheme.current) DialogContainerColorDark else DialogContainerColorLight

val ColorScheme.dialogIndicateColor: Color
    @Composable get() = if (LocalAppDarkTheme.current) DialogIndicateColorDark else DialogIndicateColorLight

val ColorScheme.dialogPlaceholderColor: Color
    @Composable get() = if (LocalAppDarkTheme.current) DialogPlaceholderDark else DialogPlaceholderLight

// Popup & Menu
val ColorScheme.popupBackground: Color
    @Composable get() = if (LocalAppDarkTheme.current) PopupBackgroundDark else PopupBackgroundLight

val ColorScheme.dropdownLabel: Color
    @Composable get() = if (LocalAppDarkTheme.current) DropdownLabelDark else DropdownLabelLight

val ColorScheme.dropdownSub: Color
    @Composable get() = if (LocalAppDarkTheme.current) DropdownSubDark else DropdownSubLight

val ColorScheme.dropdownCheck: Color
    @Composable get() = if (LocalAppDarkTheme.current) DropdownCheckDark else DropdownCheckLight

val ColorScheme.dropDivider: Color
    @Composable get() = if (LocalAppDarkTheme.current) DividerDark else DividerLight

// Chat Screen
val ColorScheme.chatStatusBarBg: Color
    @Composable get() = if (LocalAppDarkTheme.current) ChatStatusBarBgDark else ChatStatusBarBgLight

val ColorScheme.ChatTitleBarIcon: Color
    @Composable get() = if (LocalAppDarkTheme.current) ChatTitleBarIconDark else ChatTitleBarIconLight

val ColorScheme.chatBg: Color
    @Composable get() = if (LocalAppDarkTheme.current) ChatBgDark else ChatBgLight

val ColorScheme.chatDivider: Color
    @Composable get() = if (LocalAppDarkTheme.current) ChatDividerDark else ChatDividerLight

val ColorScheme.chatBtnBorder: Color
    @Composable get() = if (LocalAppDarkTheme.current) ChatBtnBorderDark else ChatBtnBorderLight

val ColorScheme.chatBtnText: Color
    @Composable get() = if (LocalAppDarkTheme.current) ChatBtnTextDark else ChatBtnTextLight

val ColorScheme.chatInfoText: Color
    @Composable get() = if (LocalAppDarkTheme.current) ChatInfoTextDark else ChatInfoTextLight

val ColorScheme.progressBarColor: Color
    @Composable get() = if (LocalAppDarkTheme.current) ProgressBarColorDark else ProgressBarColorLight

val ColorScheme.progressTrackColor: Color
    @Composable get() = if (LocalAppDarkTheme.current) ProgressTrackColorDark else ProgressTrackColorLight

// Chat List Popup
val ColorScheme.popupBg: Color
    @Composable get() = if (LocalAppDarkTheme.current) PopupBgDark else PopupBgLight

val ColorScheme.popupMask: Color
    @Composable get() = if (LocalAppDarkTheme.current) PopupMaskDark else PopupMaskLight

val ColorScheme.menuRowText: Color
    @Composable get() = if (LocalAppDarkTheme.current) MenuRowTextDark else MenuRowTextLight

val ColorScheme.menuRowIcon: Color
    @Composable get() = if (LocalAppDarkTheme.current) MenuRowIconDark else MenuRowIconLight

// Chat Input
val ColorScheme.chatInputBackground: Color
    @Composable get() = if (LocalAppDarkTheme.current) ChatInputBackgroundDark else ChatInputBackgroundLight

val ColorScheme.placeholder: Color
    @Composable get() = if (LocalAppDarkTheme.current) PlaceholderDark else PlaceholderLight

val ColorScheme.chatInputBorder: Color
    @Composable get() = if (LocalAppDarkTheme.current) chatInputBorderDark else chatInputBorderLight

// Chat List
val ColorScheme.chatLogoText: Color
    @Composable get() = if (LocalAppDarkTheme.current) ChatLogoTextDark else ChatLogoTextLight

val ColorScheme.chatMessageThinkBg: Color
    @Composable get() = if (LocalAppDarkTheme.current) ChatMessageThinkBgDark else ChatMessageThinkBgLight

val ColorScheme.chatMessageThinkBorder: Color
    @Composable get() = if (LocalAppDarkTheme.current) ChatMessageThinkBorderDark else ChatMessageThinkBorderLight

val ColorScheme.chatMessageThinkTitle: Color
    @Composable get() = if (LocalAppDarkTheme.current) ChatMessageThinkTitleDark else ChatMessageThinkTitleLight

val ColorScheme.chatMessageThinkExpandIcon: Color
    @Composable get() = if (LocalAppDarkTheme.current) ChatMessageThinkExpandIconDark else ChatMessageThinkExpandIconLight

val ColorScheme.chatMessageThinkContent: Color
    @Composable get() = if (LocalAppDarkTheme.current) ChatMessageThinkContentDark else ChatMessageThinkContentLight

val ColorScheme.chatMessageUserBg: Color
    @Composable get() = if (LocalAppDarkTheme.current) ChatMessageUserBgDark else ChatMessageUserBgLight

val ColorScheme.chatMessageUserText: Color
    @Composable get() = if (LocalAppDarkTheme.current) ChatMessageUserTextDark else ChatMessageUserTextLight

val ColorScheme.chatMessageAssistantText: Color
    @Composable get() = if (LocalAppDarkTheme.current) ChatMessageAssistantTextDark else ChatMessageAssistantTextLight

val ColorScheme.chatMessageAssistantProfiling: Color
    @Composable get() = if (LocalAppDarkTheme.current) ChatMessageAssistantProfilingDark else ChatMessageAssistantProfilingLight

val ColorScheme.ChatMessageAssistantProfilingButton: Color
    @Composable get() = if (LocalAppDarkTheme.current) ChatMessageAssistantProfilingButtonDark else ChatMessageAssistantProfilingButtonLight

val ColorScheme.chatMessageAssistantInfoTitle: Color
    @Composable get() = if (LocalAppDarkTheme.current) ChatMessageAssistantInfoTitleDark else ChatMessageAssistantInfoTitleLight

val ColorScheme.chatMessageAssistantInfoValue: Color
    @Composable get() = if (LocalAppDarkTheme.current) ChatMessageAssistantInfoValueDark else ChatMessageAssistantInfoValueLight


val ColorScheme.chatMessageAssistantInfoUnit: Color
    @Composable get() = if (LocalAppDarkTheme.current) ChatMessageAssistantInfoUnitDark else ChatMessageAssistantInfoUnitLight

val ColorScheme.chatMessageAudioBg: Color
    @Composable get() = if (LocalAppDarkTheme.current) ChatMessageAudioBgDark else ChatMessageAudioBgLight

val ColorScheme.chatMessageAudioBorder: Color
    @Composable get() = if (LocalAppDarkTheme.current) ChatMessageAudioBorderDark else ChatMessageAudioBorderLight

val ColorScheme.chatMessageAudioIcon: Color
    @Composable get() = if (LocalAppDarkTheme.current) ChatMessageAudioIconDark else ChatMessageAudioIconLight

val ColorScheme.chatMessageAudioFilename: Color
    @Composable get() = if (LocalAppDarkTheme.current) ChatMessageAudioFilenameDark else ChatMessageAudioFilenameLight

val ColorScheme.chatMessageAudioTime: Color
    @Composable get() = if (LocalAppDarkTheme.current) ChatMessageAudioTimeDark else ChatMessageAudioTimeLight

val ColorScheme.chatMessageAudioProgress: Color
    @Composable get() = if (LocalAppDarkTheme.current) ChatMessageAudioProgressDark else ChatMessageAudioProgressLight

val ColorScheme.chatMessageAudioTrack: Color
    @Composable get() = if (LocalAppDarkTheme.current) ChatMessageAudioTrackDark else ChatMessageAudioTrackLight

val ColorScheme.chatMessageAudioPlay: Color
    @Composable get() = if (LocalAppDarkTheme.current) ChatMessageAudioPlayDark else ChatMessageAudioPlayLight

val ColorScheme.chatMessageAudioPause: Color
    @Composable get() = if (LocalAppDarkTheme.current) ChatMessageAudioPauseDark else ChatMessageAudioPauseLight

val ColorScheme.chatMessageInfoCardBorder: Color
    @Composable get() = if (LocalAppDarkTheme.current) ChatMessageInfoCardBorderDark else ChatMessageInfoCardBorderLight

val ColorScheme.chatMessageFullscreenBg: Color
    @Composable get() = if (LocalAppDarkTheme.current) ChatMessageFullscreenBgDark else ChatMessageFullscreenBgLight

val ColorScheme.chatMessageDivider: Color
    @Composable get() = if (LocalAppDarkTheme.current) ChatMessageDividerDark else ChatMessageDividerLight

val ColorScheme.chatImagePreviewBg: Color
    @Composable get() = if (LocalAppDarkTheme.current) ChatImagePreviewBgDark else ChatImagePreviewBgLight

val ColorScheme.chatImagePreviewTitle: Color
    @Composable get() = if (LocalAppDarkTheme.current) ChatImagePreviewTitleDark else ChatImagePreviewTitleLight

val ColorScheme.chatImagePreviewDivider: Color
    @Composable get() = if (LocalAppDarkTheme.current) ChatImagePreviewDividerDark else ChatImagePreviewDividerLight

// Model Download List
val ColorScheme.modelDownloadListBg: Color
    @Composable get() = if (LocalAppDarkTheme.current) ModelDownloadListBgDark else ModelDownloadListBgLight

val ColorScheme.modelDownloadListSectionBg: Color
    @Composable get() = if (LocalAppDarkTheme.current) ModelDownloadListSectionBgDark else ModelDownloadListSectionBgLight


val ColorScheme.modelButtonVector: Color
    @Composable get() = if (LocalAppDarkTheme.current) modelButtonVectorDark else modelButtonVectorLight
val ColorScheme.modelButtonDisableVector: Color
    @Composable get() = if (LocalAppDarkTheme.current) modelButtonDisableVectorDark else modelButtonDisableVectorLight

// Model Download List
val ColorScheme.modelDownloadListBorder: Color
    @Composable get() = if (LocalAppDarkTheme.current) ModelDownloadListBorderDark else ModelDownloadListBorderLight

val ColorScheme.modelDownloadListTagBg: Color
    @Composable get() = if (LocalAppDarkTheme.current) ModelDownloadListTagBgDark else ModelDownloadListTagBgLight

val ColorScheme.modelDownloadListTagText: Color
    @Composable get() = if (LocalAppDarkTheme.current) ModelDownloadListTagTextDark else ModelDownloadListTagTextLight

val ColorScheme.modelDownloadListCategoryText: Color
    @Composable get() = if (LocalAppDarkTheme.current) ModelDownloadListCategoryTextDark else ModelDownloadListCategoryTextLight

// Model Card
val ColorScheme.modelCardBg: Color
    @Composable get() = if (LocalAppDarkTheme.current) ModelCardBgDark else ModelCardBgLight

val ColorScheme.modelCardTitle: Color
    @Composable get() = if (LocalAppDarkTheme.current) ModelCardTitleDark else ModelCardTitleLight

val ColorScheme.modelCardDesc: Color
    @Composable get() = if (LocalAppDarkTheme.current) ModelCardDescDark else ModelCardDescLight

val ColorScheme.modelCardTagBorder: Color
    @Composable get() = if (LocalAppDarkTheme.current) ModelCardTagBorderDark else ModelCardTagBorderLight

// Action Button
val ColorScheme.modelActionButtonBorder: Color
    @Composable get() = if (LocalAppDarkTheme.current) ModelActionButtonBorderDark else ModelActionButtonBorderLight

val ColorScheme.modelActionButtonRunBorder: Color
    @Composable get() = if (LocalAppDarkTheme.current) ModelActionButtonRunBorderDark else ModelActionButtonRunBorderLight

val ColorScheme.modelActionButtonRunContent: Color
    @Composable get() = if (LocalAppDarkTheme.current) ModelActionButtonRunContentDark else ModelActionButtonRunContentLight

val ColorScheme.modelActionButtonDisableRunBorder: Color
    @Composable get() = if (LocalAppDarkTheme.current) ModelActionButtonRunDisableBorderDark else ModelActionButtonDisableRunBorderLight

val ColorScheme.modelActionButtonRunDisableContent: Color
    @Composable get() = if (LocalAppDarkTheme.current) ModelActionButtonRunDisableContentDark else ModelActionButtonRunDisableContentLight

val ColorScheme.modelActionButtonDownloadBg: Color
    @Composable get() = if (LocalAppDarkTheme.current) ModelActionButtonDownloadBgDark else ModelActionButtonDownloadBgLight

val ColorScheme.modelActionButtonDownloadText: Color
    @Composable get() = if (LocalAppDarkTheme.current) ModelActionButtonDownloadTextDark else ModelActionButtonDownloadTextLight

// Download Progress
val ColorScheme.modelDownloadProgressBar: Color
    @Composable get() = if (LocalAppDarkTheme.current) ModelDownloadProgressBarDark else ModelDownloadProgressBarLight

val ColorScheme.modelDownloadProgressText: Color
    @Composable get() = if (LocalAppDarkTheme.current) ModelDownloadProgressTextDark else ModelDownloadProgressTextLight

val ColorScheme.modelDownloadProgressTrack: Color
    @Composable get() = if (LocalAppDarkTheme.current) ModelDownloadProgressTrackDark else ModelDownloadProgressTrackLight

val ColorScheme.modelDownloadScreenBg: Color
    @Composable
    get() = if (LocalAppDarkTheme.current) modelDownloadScreenBgDark else modelDownloadScreenBgLight

val ColorScheme.modelDownloadDivider: Color
    @Composable
    get() = if (LocalAppDarkTheme.current) modelDownloadDividerDark else modelDownloadDividerLight

val ColorScheme.modelTopBarBg: Color
    @Composable
    get() = if (LocalAppDarkTheme.current) modelTopBarBgDark else modelTopBarBgLight

val ColorScheme.modelTopBarTitle: Color
    @Composable
    get() = if (LocalAppDarkTheme.current) modelTopBarTitleDark else modelTopBarTitleLight

val ColorScheme.settingBg: Color
    @Composable get() = if (LocalAppDarkTheme.current) settingBgDark else settingBgLight

val ColorScheme.settingTextColor: Color
    @Composable get() = if (LocalAppDarkTheme.current) settingTextColorDark else settingTextColorLight

val ColorScheme.settingSubTextColor: Color
    @Composable get() = if (LocalAppDarkTheme.current) settingSubTextColorDark else settingSubTextColorLight

val ColorScheme.settingDivider: Color
    @Composable get() = if (LocalAppDarkTheme.current) settingDividerDark else settingDividerLight

val ColorScheme.settingBtnBorder: Color
    @Composable get() = if (LocalAppDarkTheme.current) settingBtnBorderDark else settingBtnBorderLight

val ColorScheme.settingBtnBg: Color
    @Composable get() = if (LocalAppDarkTheme.current) settingBtnBgDark else settingBtnBgLight

val ColorScheme.settingBtnText: Color
    @Composable get() = if (LocalAppDarkTheme.current) settingBtnTextDark else settingBtnTextLight


val ColorScheme.settingSliderActive: Color
    @Composable get() = if (LocalAppDarkTheme.current) settingSliderActiveDark else settingSliderActiveLight

val ColorScheme.settingSliderInactive: Color
    @Composable get() = if (LocalAppDarkTheme.current) settingSliderInactiveDark else settingSliderInactiveLight

val ColorScheme.settingSliderThumb: Color
    @Composable get() = if (LocalAppDarkTheme.current) settingSliderThumbDark else settingSliderThumbLight

val ColorScheme.textFieldBg: Color
    @Composable get() = if (LocalAppDarkTheme.current) textFieldBgDark else textFieldBgLight

val ColorScheme.textFieldBorder: Color
    @Composable get() = if (LocalAppDarkTheme.current) textFieldBorderDark else textFieldBorderLight

val ColorScheme.textFieldText: Color
    @Composable get() = if (LocalAppDarkTheme.current) textFieldTextDark else textFieldTextLight

val ColorScheme.textFieldTextDisabled: Color
    @Composable get() = if (LocalAppDarkTheme.current) textFieldTextDisabledDark else textFieldTextDisabledLight

val ColorScheme.switchThumb: Color
    @Composable get() = if (LocalAppDarkTheme.current) switchThumbDark else switchThumbLight

val ColorScheme.switchTrackChecked: Color
    @Composable get() = if (LocalAppDarkTheme.current) switchTrackCheckedDark else switchTrackCheckedLight

val ColorScheme.switchTrackUnchecked: Color
    @Composable get() = if (LocalAppDarkTheme.current) switchTrackUncheckedDark else switchTrackUncheckedLight

val ColorScheme.textScrollFieldBg: Color
    @Composable get() = if (LocalAppDarkTheme.current) textFieldScrollBgDark else textFieldScrollBgLight

val ColorScheme.textScrollFieldText: Color
    @Composable get() = if (LocalAppDarkTheme.current) textFieldScrollTextDark else textFieldScrollTextLight

val ColorScheme.SettingIcon: Color
    @Composable get() = if (LocalAppDarkTheme.current) SettingIconDark else SettingIconLight

val ColorScheme.generationSettingsBg: Color
    @Composable get() = if (LocalAppDarkTheme.current) GenerationSettingsBgDark else GenerationSettingsBgLight

val ColorScheme.generationSettingsHeader: Color
    @Composable get() = if (LocalAppDarkTheme.current) GenerationSettingsHeaderDark else GenerationSettingsHeaderLight

val ColorScheme.generationSettingsDivider: Color
    @Composable get() = if (LocalAppDarkTheme.current) GenerationSettingsDividerDark else GenerationSettingsDividerLight

val ColorScheme.generationSettingsButton: Color
    @Composable get() = if (LocalAppDarkTheme.current) GenerationSettingsButtonDark else GenerationSettingsButtonLight

val ColorScheme.generationSettingsButtonBg: Color
    @Composable get() = if (LocalAppDarkTheme.current) GenerationSettingsButtonBgDark else GenerationSettingsButtonBgLight

val ColorScheme.generationSettingsAdvancedArrow: Color
    @Composable get() = if (LocalAppDarkTheme.current) GenerationSettingsAdvancedArrowDark else GenerationSettingsAdvancedArrowLight

val ColorScheme.commonSettingsSegmentedBg: Color
    @Composable get() = if (LocalAppDarkTheme.current) CommonSettingsSegmentedBgDark else CommonSettingsSegmentedBgLight

val ColorScheme.commonSettingsSelectedText: Color
    @Composable get() = if (LocalAppDarkTheme.current) CommonSettingsSelectedTextDark else CommonSettingsSelectedTextLight

val ColorScheme.commonSettingsUnSelectedText: Color
    @Composable get() = if (LocalAppDarkTheme.current) CommonSettingsUnSelectedTextDark else CommonSettingsUnSelectedTextLight

val ColorScheme.commonSettingsSelectedBg: Color
    @Composable get() = if (LocalAppDarkTheme.current) CommonSettingsSelectedBgDark else CommonSettingsSelectedBgLight

val ColorScheme.appInfoText: Color
    @Composable get() = if (LocalAppDarkTheme.current) AppInfoTextColorDark else AppInfoTextColorLight

val ColorScheme.errorBannerBorder: Color
    @Composable get() = if (LocalAppDarkTheme.current) ErrorBannerBorderColorDark else ErrorBannerBorderColorLight

val ColorScheme.errorBannerBg: Color
    @Composable get() = if (LocalAppDarkTheme.current) ErrorBannerBgColorDark else ErrorBannerBgColorLight

val ColorScheme.curModelTagBg: Color
    @Composable get() = if (LocalAppDarkTheme.current) CurModelTagBgColorDark else CurModelTagBgColorLight

val ColorScheme.curModelTagTextColor: Color
    @Composable get() = if (LocalAppDarkTheme.current) CurModelTagTextColorDark else CurModelTagTextColorLight

val ColorScheme.curModelTagTintColor: Color
    @Composable get() = if (LocalAppDarkTheme.current) CurModelTagTintColorDark else CurModelTagTintColorLight

val ColorScheme.errorBannerText: Color
    @Composable get() = if (LocalAppDarkTheme.current) ErrorBannerTextColorDark else ErrorBannerTextColorLight

val ColorScheme.editMessageBgColor: Color
    @Composable get() = if (LocalAppDarkTheme.current) EditMessageBgColorDark else EditMessageBgColorLight

val ColorScheme.editMessageTextColor: Color
    @Composable get() = if (LocalAppDarkTheme.current) EditMessageTextColorDark else EditMessageTextColorLight

val ColorScheme.editMessagePenColor: Color
    @Composable get() = if (LocalAppDarkTheme.current) EditMessagePenColorDark else EditMessagePenColorLight

val ColorScheme.editMessageCancelColor: Color
    @Composable get() = if (LocalAppDarkTheme.current) EditMessageCancelColorDark else EditMessageCancelColorLight
// Splash
val ColorScheme.handleColor: Color
    @Composable get() = if (LocalAppDarkTheme.current) SplashBackgroundDark else Primary

val ColorScheme.audioWaveLineColor: Color
    @Composable get() = if (LocalAppDarkTheme.current) Color_323232 else Color_e7e7e7
val ColorScheme.audioWaveColor: Color
    @Composable get() = if (LocalAppDarkTheme.current) Color_e7e7e7 else Color_454545
