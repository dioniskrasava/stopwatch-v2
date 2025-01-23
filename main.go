package main

import (
	"fmt"
	"image/color"
	"log"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// ТЕСТ ГИТА
// ВТОРОЙ ТЕСТ

// СТРУКТУРА ОПИСЫВАЮЩАЯ ФЛАГИ ПРИЛОЖЕНИЯ СЕКУНДОМЕРА
type flags_app struct {
	start_pressed bool
	paused        bool
	begin_flag    bool
}

var flags flags_app = flags_app{false, false, true}

// структура кнопок
type widgets_app struct {
	btn_str         *widget.Button
	btn_pause       *widget.Button
	btn_clear       *widget.Button
	input           *canvas.Text
	cont_begin      *fyne.Container
	w               fyne.Window
	btn_cont        *fyne.Container
	ent_activity    *widget.Select
	cont_add_act    *fyne.Container
	input_act       *widget.Entry
	input_act_descr *widget.Entry
}

var wdgts widgets_app

// структура кол-ва времени в таймере
type time_values struct {
	h           int
	m           int
	s           int
	ms          int
	msTimer     int
	msStartUnix int
	timePause   int
	times       string
}

var time_val time_values = time_values{
	times: "00:00:00:000",
}

type information_app struct {
	w_a float32
	h_a float32
	DB  *gorm.DB
}

var inf_app information_app

type Activity struct {
	Id            uint
	Name_activity string
	Description   string
}

// Приложение секундомера
func stopwatch() {
	/*Тут будет находится всё касаемо окна приложения*/

	inf_app.w_a = 350 // ширина
	inf_app.h_a = 120 // высота

	a := app.New()
	wdgts.w = a.NewWindow("Stopwatch")
	wdgts.w.Resize(fyne.NewSize(inf_app.w_a, inf_app.h_a)) // Увеличение размера окна
	//w.SetFixedSize(true)
	wdgts.w.CenterOnScreen()

	//var wdgts widgets_app // для получения начальных виджетов
	widgets_begin()

	wdgts.w.SetContent(wdgts.cont_begin)
	wdgts.w.SetMainMenu(init_menu_stpwtc()) // и передача виджетов ф-ии по инициал. меню
	wdgts.w.ShowAndRun()

}

// инициализация НАЧАЛЬНОГО СОСТОЯНИЯ секундомера
func widgets_begin() {

	input := canvas.NewText("00:00:00:000", color.White)
	input.TextSize = 35 // Увеличение размера шрифта

	btn_str := widget.NewButton("       Start       ", nil)
	btn_pause := widget.NewButton("      Pause      ", nil)
	btn_clear := widget.NewButton("      Clear      ", nil)

	// инициализируем массив кнопок для передачи в ф-ю logic_timer

	//var btns witgets_app = witgets_app{btn_str, btn_pause, btn_clear, input}
	// теперь приходится инициализировать объект нашей суперструктуры частями и в таком
	// формате. Потому что не все поля структуры известны на момент инициализации
	// а взаимодействовать уже необходимо
	wdgts.btn_str = btn_str
	wdgts.btn_pause = btn_pause
	wdgts.btn_clear = btn_clear
	wdgts.input = input

	// изначально кнопки паузы и чистки - неактивны
	btn_pause.Disable()
	btn_clear.Disable()

	btn_str.OnTapped = logic_timer("start", &flags, wdgts, &time_val)
	btn_pause.OnTapped = logic_timer("pause", &flags, wdgts, &time_val)
	btn_clear.OnTapped = logic_timer("clear", &flags, wdgts, &time_val)

	wdgts.btn_cont = container.NewHBox(btn_str, btn_pause, btn_clear)

	wdgts.cont_begin = container.NewBorder(
		nil,
		container.NewVBox(container.NewCenter(input), container.NewCenter(wdgts.btn_cont)),
		nil,
		nil,
		nil,
	)
}

// логика работы таймера
func logic_timer(s string, f *flags_app, w widgets_app, t *time_values) func() {
	// функция обновления состояния кнопок
	updateButtonsState := func(startPressed bool, paused bool) {
		w.btn_str.Disable()
		w.btn_pause.Disable()
		w.btn_clear.Disable()

		if startPressed && !paused {
			w.btn_pause.Enable()
			w.btn_clear.Enable()
		} else if !startPressed && paused {
			w.btn_str.Enable()
			w.btn_clear.Enable()
		} else if !startPressed && !paused {
			w.btn_str.Enable()
		}
	}

	// СТАРТ
	start := func() {
		if !f.start_pressed {
			f.start_pressed = true
			w.btn_clear.Enable()

			if f.paused {
				t.msStartUnix = int(time.Now().UnixMilli()) - t.timePause
			}

			f.paused = false

			if f.begin_flag {
				t.msStartUnix = int(time.Now().UnixMilli()) // время начала отсчета
				f.begin_flag = false
			}

			go func(f *flags_app) {

				for {
					time.Sleep(time.Millisecond * 50)
					if !f.paused {
						t.msTimer = int(time.Now().UnixMilli()) - t.msStartUnix // пройденное время
						t.ms = t.msTimer % 1000
						t.s = t.msTimer / 1000
						t.m = t.s / 60
						t.h = t.m / 60
						t.s = (t.msTimer - (t.h * 3600000) - (t.m * 60000) - t.ms) / 1000
						t.times = fmt.Sprintf("%02d:%02d:%02d:%03d", t.h, t.m, t.s, t.ms)
						w.input.Text = t.times
						w.input.Refresh() // Добавляем этот вызов для обновления виджета
					}

					if f.paused {
						break
					}

				}

			}(f)
		}

		updateButtonsState(f.start_pressed, f.paused)
	}

	// ПАУЗА
	pause := func() {
		f.start_pressed = false
		f.paused = true
		updateButtonsState(f.start_pressed, f.paused)
		//фиксируем время остановки в мс
		t.timePause = t.msTimer
	}

	// ОЧИСТКА ПОЛЯ
	reset := func() {
		w.input.Text = "00:00:00:000"
		w.input.Refresh()                           // Добавляем этот вызов для обновления виджета
		t.msStartUnix = int(time.Now().UnixMilli()) // это нужно если кнопка очистки нажата, но не нажата пауза. Тогда происходит очистка и продолжается счет
		f.begin_flag = true                         // на случай очищения на паузе (для запуска нового отсчета при нажатии старта)
		if f.paused {
			w.btn_clear.Disable()
		}

	}

	if s == "start" {
		return start
	} else if s == "pause" {
		return pause
	}

	// если вызывающий объект ничего не написал в параметре s то на выходе получит функцию стирания
	return reset

}

func init_menu_stpwtc() *fyne.MainMenu {
	//------------------------------------------------------------------
	file_item1 := fyne.NewMenuItem("Сохранить", func() {})
	file_item2 := fyne.NewMenuItem("Открыть", func() {})

	file_menu := fyne.NewMenu("Файл", file_item1, file_item2)
	//------------------------------------------------------------------
	view_item1 := fyne.NewMenuItem("Секундомер", func() {
		wdgts.w.SetContent(wdgts.cont_begin)                   // восстановление первоначального вида секундомера
		wdgts.w.Resize(fyne.NewSize(inf_app.w_a, inf_app.h_a)) // восстановление размера окна    !!!! ПОЧЕМУ-ТО НЕ СРАБАТЫВАЕТ КОРРЕКТНО
	})
	view_item2 := fyne.NewMenuItem("Показать активность", func() {
		begin_work_db()
		wdgts.w.Resize(fyne.NewSize(inf_app.w_a, inf_app.h_a+125)) // Увеличение размера окна
		//wdgts.ent_activity.Show()
		wdgts.w.SetContent(container.NewVBox(
			container.NewCenter(container.NewGridWrap(fyne.NewSize(inf_app.w_a, 40), wdgts.ent_activity)),
			container.NewCenter(wdgts.input),
			container.NewCenter(wdgts.btn_cont)))
	})

	view_menu := fyne.NewMenu("Вид", view_item1, view_item2)
	//------------------------------------------------------------------

	edit_item1 := fyne.NewMenuItem("Добавить активность", func() {
		init_cont_add_act()
		wdgts.w.SetContent(wdgts.cont_add_act)
	})

	edit_menu := fyne.NewMenu("Редактирование", edit_item1)
	//------------------------------------------------------------------

	main_menu := fyne.NewMainMenu(file_menu, view_menu, edit_menu)

	return main_menu

}

func init_cont_add_act() {
	// КОНТЕНТ ДОБАВЛЕНИЯ АКТИВНОСТИ
	wdgts.input_act = widget.NewEntry()
	wdgts.input_act.SetPlaceHolder("Введите название активности...")
	wdgts.input_act_descr = widget.NewEntry()
	wdgts.input_act_descr.SetPlaceHolder("Описание активности...")
	//adding_an_activity_pad := func() { return adding_an_activity() }
	btn_add_act := widget.NewButton("Добавить активность", adding_an_activity) // ? добавлять ли в общую структуру виджетов?
	wdgts.cont_add_act = container.NewVBox(wdgts.input_act, wdgts.input_act_descr, btn_add_act)

	//adding_an_activity()

}

func begin_work_db() {

	//-------------------------------------------------------------------
	//

	// инициализация видов активностей
	var names_activites []string

	result := inf_app.DB.Model(&Activity{}).Pluck("Name_activity", &names_activites)
	if result.Error != nil {
		log.Fatal("Ошибка при получении имен:", result.Error)
	}

	wdgts.ent_activity = widget.NewSelect(
		names_activites,

		func(s string) {
			fmt.Printf("Selected is %s", s)
		},
	)
	wdgts.ent_activity.PlaceHolder = "Выберите активность"
	wdgts.ent_activity.Hide()
	//------------------------------------------------------------------

	//срез видов активностей
	var activites []Activity

	//БЛОК ИНИЦИАЛИЗАЦИИ БАЗЫ ДАННЫХ
	var err error
	inf_app.DB, err = gorm.Open(sqlite.Open("todo.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("Не удалось подключиться к базе данных:", err)
	}
	inf_app.DB.AutoMigrate(&Activity{})
	inf_app.DB.Find(&activites)

}

func adding_an_activity() {
	added_act := Activity{
		Name_activity: wdgts.input_act.Text,
		Description:   wdgts.input_act_descr.Text,
	}

	inf_app.DB.Create(&added_act)
	inf_app.DB.Find(&added_act)

	// после добавления активности - ворачиваем контент секундомера
	wdgts.w.SetContent(wdgts.cont_begin)                   // восстановление первоначального вида секундомера
	wdgts.w.Resize(fyne.NewSize(inf_app.w_a, inf_app.h_a)) // восстановление размера окна
}

// больше ничего тут не должно быть
func main() {
	stopwatch()
}
