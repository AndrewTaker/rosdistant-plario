> [!WARNING]  
> Это proof of concept репозиторий. Использовать только для тестирования. Призываю решать задания самим.

# Proof of concept росидистант + пларио авто квиз

## Компиляция
Обязательные зависимости:
- [golang>=1.24.11](https://go.dev/)

Необязательные зависимости
- [Make](https://en.wikipedia.org/wiki/Make_(software))

Собрать через Makefile
```bash
cd rosdistant-plario
make
```
Собрать через go build
```
cd rosdistant-plario
go build -o ./bin/plario ./apps/cli
```

## Использование
Доступные флаги
```
  -help string
    	print out usage

  -ptoken string
    	required: plario access token
  -gtoken string
    	required: groq api token

  -subject int
    	required if not infomode: subject_id
  -course int
    	required if not infomode: course_id
  -module int
    	required if not infomode: module_id

  -infomode
    	optional: print out availabe subjects, courses, modules and exit with 0

  -loglevel string
    	optional: provide to change log level (default "info")
  -model value
    	optional: choose from available groq models (default openai/gpt-oss-120b)
  -rmax int
    	optional: set maximum value for random delay between each question submission (default 10)
  -rmin int
    	optional: set minimum value for random delay between each question submission (default 5)
  -till_mastery float
    	optional: provide if you want to stop program execution at certain mastery level float 2.f
```

Флаги подробнее:
- `ptoken` REQUIRED Access Token аккаунта plario. Можно найти в devtools -> local storage
- `gtoken` REQUIRED Api Token аккаунта Groq. Нужно получить ![на странице Groq](https://groq.com/). У них невероятно щедрый фри тир (не нунжо привязывать способ оплаты)
- `infomode` OPTIONAL Если передать флаг - получите информацию о доступных предметах, курсах и модулях в виде таблицы и выйти (`s_id` - идентификатор предмета, `c_id` - идентификатор курса, `m_id` - идентификатор модуля)
<img width="2341" height="322" alt="image" src="https://github.com/user-attachments/assets/071fa622-3de5-496c-84b0-6a9bb686d196" />

- `subject` REQURED IF NOT INFOMODE Идентификатор предмета
- `course` REQURED IF NOT INFOMODE Идентификатор курса
- `module` REQURED IF NOT INFOMODE Идентификатор модуля
- `loglevel` OPTIONAL Установить уровень логгирования default - `info`
- `model` OPTIONAL Какую модель использовать для Groq. [Информация о моделях](https://console.groq.com/docs/models)
- `rmax` OPTIONAL Максимальное время (секунды) ожидания между итерациями default - `5`
- `rmin` OPTIONAL Минимальное время (секунды) ожидания между итерациями default - `10`
- `till_mastery` OPTIONAL Установить порог мастерства на модуль, принимает число с плавающей точкой с точностью до двух знаков после запятой
- `help` OPTIONAL Вывести список флагов и выйти

## Примеры

> [!TIP]
> Токены будут в переменных окружения

> [!IMPORTANT]
> При первом запуске нужно включить `infomode`, чтобы получить доступные предметы, курсы и модули, если заранее идентификаторы не известны


```bash
export PLARIO_TOKEN=<access_token>
export GROQ_TOKEN=<groq_token>
```

Получить список доступных предметов, курсов и модулей
```bash
./bin/plario -ptoken $PLARIO_TOKEN -gtoken $GROQ_TOKEN -infomode
```

Выполнить задания по предмету `10`, курсу `1` и модулю `44`
```bash
./bin/plario -ptoken $PLARIO_TOKEN -gtoken $GROQ_TOKEN -subject 10 -course 1 -module 44
```

Выполнить задания по предмету `10`, курсу `1` и модулю `44` до тех пор, пока `mastery` по модулю не достигнет `88%` (0.88)
```bash
./bin/plario -ptoken $PLARIO_TOKEN -gtoken $GROQ_TOKEN -subject 10 -course 1 -module 44 -till_mastery 0.88
```

Выполнить задания по предмету `10`, курсу `1` и модулю `44` до тех пор, пока `mastery` по модулю не достигнет `88%` (0.88) с перерывами между запросами [10, 20], не [10, 20).
```bash
./bin/plario -ptoken $PLARIO_TOKEN -gtoken $GROQ_TOKEN -subject 10 -course 1 -module 44 -till_mastery 0.88 -rmin 10 -rmax 20
```
