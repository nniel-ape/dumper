package i18n

var ru = map[MsgKey]string{
	// Welcome and help
	MsgWelcome: `👋 <b>Добро пожаловать в Dumper!</b>

Я помогу вам собирать и организовывать знания из интернета.

<b>Как использовать:</b>
• Отправьте мне любую ссылку - я извлеку контент, создам резюме и теги
• Отправьте текстовые заметки - я тоже их категоризирую
• Используйте /search для поиска сохранённых записей
• Используйте /recent для просмотра последних записей
• Используйте /tags для просмотра всех тегов
• Используйте /lang для смены языка

Все ваши данные хранятся приватно и могут быть экспортированы в Obsidian.`,

	MsgHelp: `<b>Команды:</b>
/search [запрос] - Поиск по сохранённым записям
/recent - Показать последние записи
/tags - Список всех тегов
/stats - Статистика хранилища
/export - Экспорт в формат Obsidian
/app - Открыть Mini App (если настроен)
/lang - Сменить язык (en/ru)

<b>Сохранение контента:</b>
Просто отправьте мне любую ссылку или текстовое сообщение!`,

	MsgUnknownCommand: "Неизвестная команда. Используйте /help для просмотра доступных команд.",

	// Processing status
	MsgProcessingLink:   "⏳ Обрабатываю ссылку...",
	MsgProcessingNote:   "⏳ Обрабатываю заметку...",
	MsgSavingImage:      "📷 Сохраняю изображение...",
	MsgSavingImages:     "📷 Сохраняю %d изображений...",
	MsgSearching:        "🔍 Ищу: <b>%s</b>...",
	MsgSearchUsage:      "Использование: /search [запрос]\nПример: /search golang concurrency",
	MsgRecentItems:      "📚 <b>Последние записи:</b>\n\n",
	MsgYourTags:         "🏷 <b>Ваши теги:</b>\n\n#%s",
	MsgYourVault:        "📊 <b>Ваше хранилище:</b>\n\n• Записей: %d\n• Тегов: %d",
	MsgOpenApp:          "📱 Открыть приложение",
	MsgViewInApp:        "Открыть в приложении",
	MsgAppNotConfigured: "Mini App не настроен. Установите переменную окружения WEBAPP_URL.",
	MsgOpenMiniApp:      "Откройте Mini App для просмотра, поиска и визуализации ваших знаний:",
	MsgExportComingSoon: "Функция экспорта скоро появится! Пока используйте API /api/export.",

	// Success messages
	MsgSaved:      "✅ <b>Сохранено!</b>",
	MsgImageSaved:  "✅ <b>Изображение сохранено!</b>",
	MsgImagesSaved: "✅ <b>%d изображений сохранено!</b>",

	// Empty states
	MsgNoResults: "Ничего не найдено.",
	MsgNoItems:   "Пока нет сохранённых записей. Отправьте мне ссылку или заметку!",
	MsgNoTags:    "Пока нет тегов.",
	MsgSearchFor: `🔍 <b>Результаты по запросу "%s":</b>`,

	// Errors
	MsgFailedProcess:   "❌ Ошибка обработки: %v",
	MsgFailedVault:     "❌ Не удалось получить доступ к хранилищу",
	MsgFailedSearch:    "❌ Ошибка поиска: %v",
	MsgFailedListItems: "❌ Не удалось получить список записей: %v",
	MsgFailedGetTags:   "❌ Не удалось получить теги: %v",
	MsgFailedGetStats:  "❌ Не удалось получить статистику: %v",
	MsgFailedFileInfo:  "❌ Не удалось получить информацию о файле: %v",
	MsgFailedDownload:  "❌ Не удалось скачать изображение: %v",
	MsgFailedReadImage: "❌ Не удалось прочитать изображение: %v",
	MsgFailedSaveImage: "❌ Не удалось сохранить изображение: %v",

	// Language
	MsgLangCurrent: "🌐 Текущий язык: <b>Русский</b>\n\nИспользуйте /lang en для переключения на английский.",
	MsgLangUsage:   "Использование: /lang [en|ru]\n\nДоступные языки:\n• en - English\n• ru - Русский",
	MsgLangChanged: "✅ Язык изменён на <b>Русский</b>",
	MsgLangUnknown: "Неизвестный язык. Доступны: en, ru",
}
