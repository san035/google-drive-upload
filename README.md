# Отправка файла в Google Drive

Проект на Go для загрузки файла `send_file.txt` в Google Drive личного аккаунта.

## Установка зависимостей

```bash
go mod init drive-uploader
go get golang.org/x/oauth2/google
go get google.golang.org/api/drive/v3
```

## Получение credentials для Google Drive API

### Шаг 1: Создание проекта в Google Cloud Console

1. Перейдите на [Google Cloud Console](https://console.cloud.google.com/)
2. Нажмите **Select a project** → **New Project**
3. Введите имя проекта (например, `DriveUploader`)
4. Нажмите **Create**

### Шаг 2: Включение Google Drive API

1. В меню слева выберите **APIs & Services** → **Library**
2. В поиске введите **Google Drive API**
3. Выберите **Google Drive API** из результатов
4. Нажмите **Enable**

### Шаг 3: Настройка OAuth Consent Screen

1. В меню слева выберите **OAuth consent screen**
2. Выберите тип пользователя: **External**
3. Нажмите **Create**

4. Заполните обязательные поля:
   - **App name**: `DriveUploader` (или любое имя)
   - **User support email**: выберите ваш email `a.n.skoroh@gmail.com`
   - **Scopes**: нажмите **Add or remove scopes**, найдите `.../auth/drive.file` и выберите его
   - **Test users**: нажмите **Add users**, введите `a.n.skoroh@gmail.com`

5. Нажмите **Save and continue** → **Back to dashboard**

6. Нажмите кнопку **PUBLISH APP** и подтвердите публикацию (выберите "Make App available to all users")

### Шаг 4: Создание OAuth 2.0 credentials

1. Перейдите в **APIs & Services** → **Credentials**
2. Нажмите **+ CREATE CREDENTIALS** → **OAuth client ID**
3. Выберите тип приложения **Desktop application**
4. Введите имя (например, `DriveUploader`)
5. Нажмите **Create**
6. Скопируйте **Client ID** и **Client Secret**
7. Нажмите **Download JSON** для скачивания файла credentials

### Шаг 5: Сохранение credentials

1. Переименуйте скачанный файл в `credentials.json`
2. Поместите его в папку с программой

## Запуск программы

```bash
go run main.go
```

При первом запуске:
1. Программа выведет ссылку для авторизации
2. Перейдите по ссылке и войдите в аккаунт `a.n.skoroh@gmail.com`
3. Разрешите доступ к Google Drive
4. Скопируйте полученный код
5. Вставьте код в терминал

Файл `token.json` будет создан автоматически для последующих запусков.

## О токенах

При авторизации программа получает два токена:

| Токен | Назначение | Срок действия |
|-------|------------|---------------|
| **Access Token** | Доступ к API Drive | ~1 час |
| **Refresh Token** | Обновление Access Token | До 6-12 месяцев |

Файл `token.json` содержит оба токена. При истечении Access Token программа автоматически использует Refresh Token для получения нового — **повторная авторизация в браузере не требуется**.

### Важно:
- Если `token.json` удалить — потребуется повторная авторизация
- Refresh Token может быть отозван Google, если не использовать программу более 6 месяцев
- Периодически запускайте программу (раз в 1-2 месяца), чтобы поддерживать токен в активном состоянии

## Файлы

- `credentials.json` — OAuth credentials от Google Cloud Console
- `token.json` — сохраненный токен доступа (создается автоматически)
- `send_file.txt` — файл для загрузки