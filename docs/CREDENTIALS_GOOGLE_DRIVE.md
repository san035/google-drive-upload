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
2. Нажмите кнопку **Get started**

3. Заполните обязательные поля на странице **App Information**:
   - **App name**: `DriveUploader` (или любое имя)
   - **User support email**: выберите ваш email

6. Нажмите **Next** 

7. На странице **Audience** выберите **External**, нажмите **Next**

8. Нажмите **Next** на странице **Contact Information**

9. Подтвердите согласие: отметьте **I agree to the Google API Services: User Data Policy** и нажмите **Continue** и **Create**

10. Слева в меню выбрать Audience

11. В разделе **Test users** нажмите **Add users**

12. Введите ваш email (или email пользователя, которому нужен доступ)

13. Нажмите **Save**


## Шаг 4: Создание OAuth 2.0 credentials

1. Перейдите в **APIs & Services** → **Credentials**
2. Нажмите **+ CREATE CREDENTIALS** → **OAuth client ID**
3. Выберите тип приложения **Desktop application**
4. Введите имя (например, `DriveUploader`)
5. Нажмите **Create**
6. Нажмите **Download JSON** для скачивания файла credentials

## Шаг 5: Сохранение credentials

1. Переименуйте скачанный файл в имя которое указано в поле google_credentials_file файла config.yaml
2. Поместите его в папку с программой

## Шаг 6 Добавить email в список тестировщиков:

1. Перейдите в [Google Cloud Console](https://console.cloud.google.com/)
2. Выберите ваш проект
3. В меню слева выберите **OAuth consent screen**
