# Получение credentials для Google Drive API

Документ описывает процесс получения OAuth 2.0 credentials для работы с Google Drive API.

## Шаг 1: Создание проекта в Google Cloud Console

1. Перейдите на [Google Cloud Console](https://console.cloud.google.com/)
2. Нажмите **Select a project** → **New Project**
3. Введите имя проекта (например, `DriveUploader`)
4. Нажмите **Create**

## Шаг 2: Включение Google Drive API

1. В меню слева выберите **APIs & Services** → **Library**
2. В поиске введите **Google Drive API**
3. Выберите **Google Drive API** из результатов
4. Нажмите **Enable**

## Шаг 3: Настройка OAuth Consent Screen

1. В меню слева выберите **OAuth consent screen**
2. Выберите тип пользователя: **External**
3. Нажмите **Create**

4. Заполните обязательные поля:
   - **App name**: `DriveUploader` (или любое имя)
   - **User support email**: выберите ваш email
   - **Scopes**: нажмите **Add or remove scopes**, найдите `.../auth/drive.file` и выберите его
   - **Test users**: нажмите **Add users**, введите ваш email

5. Нажмите **Save and continue** → **Back to dashboard**

6. Нажмите кнопку **PUBLISH APP** и подтвердите публикацию (выберите "Make App available to all users")

## Шаг 4: Создание OAuth 2.0 credentials

1. Перейдите в **APIs & Services** → **Credentials**
2. Нажмите **+ CREATE CREDENTIALS** → **OAuth client ID**
3. Выберите тип приложения **Desktop application**
4. Введите имя (например, `DriveUploader`)
5. Нажмите **Create**
6. Скопируйте **Client ID** и **Client Secret**
7. Нажмите **Download JSON** для скачивания файла credentials

## Шаг 5: Сохранение credentials

1. Переименуйте скачанный файл в `credentials.json`
2. Поместите его в папку с программой