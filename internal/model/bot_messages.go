package model

const (
	helloMessage = "Привет! 👋\n\n" + currencyHelpMessage + "\n\n\n" + addHelpMessage + "\n\n\n" + reportHelpMessage

	currencyHelpMessage = `Для смены текущей валюты используй команду /currency.`

	currencyCurrentMessage = `Текущая валюта: `
	currencyChooseMessage  = `Выбери валюту:`
	currencyLaterMessage   = "🚧 Выполняется обновление курсов валют. 🚧\n\nПовтори попытку чуть позже."

	addHelpMessage = `Чтобы добавить запись о расходах, отправь команду:
<pre>
/add [дата] &lt;сумма&gt; &lt;категория&gt;
</pre>
Дата может быть указана в формате <b>dd.mm.yyyy</b> (день.месяц.год).
Чтобы задать сегодняшнее число, можно использовать знак <b>@</b> в качестве даты, или не указывать дату совсем.
Кроме того, в качестве даты можно использовать строку вида <b>-Nd</b>, где N — количество "дней назад" (1 можно не указывать).
Например, <b>-2d</b> значит "2 дня назад".

Сумма указывается в формате <b>XX[.yy]</b>: целого или дробного числа с одним или двумя знаками после запятой (вместо которой можно использовать точку).`

	reportHelpMessage = `Для просмотра расходов по категориям выполни одну из команд (w — расходы за неделю, m — за месяц, y — за год):
<pre>
/report [N]w
/report [N]m
/report [N]y
</pre>
Если задать положительное число N, будут выведены расходы за N последних недель/месяцев/лет.

Команда <code>/report</code> (без дополнительных параметров) вернёт расходы за последнюю неделю. 
`

	doneMessage = `Готово!`

	sorryMessage = "Извини, я не знаю такой команды. 🙁\n\n" + currencyHelpMessage + "\n\n\n" + addHelpMessage + "\n\n\n" + reportHelpMessage
)

func errorMessage(err error, fail, help string) string {
	return fail + "\nОшибка: " + err.Error() + ".\n\n\nДля справки:\n" + help
}