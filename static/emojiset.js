const el = new EmojiPopover({
    button: '.picker',
    targetElement: '.emoji-picker',
    emojiList: [
            {
              value: '🤣',
              label: 'Сміюсь до сліз'
            },
            {
              value: '😃',
              label: 'Великий сміх'
            },
            {
              value: '😅',
              label: 'Гіркий сміх'
            },
            {
              value: '😆',
              label: 'Похмуро сміється'
            },
            {
              value: '😏',
              label: 'Задоволений'
            },
            {
              value: '😊',
              label: 'Посмішка'
            },
            {
              value: '😎',
              label: 'Круто!'
            },
            {
              value: '😍',
              label: 'Вмираю від захоплення'
            },
            {
              value: '🙂',
              label: 'Ха-ха'
            },
            {
              value: '🤩',
              label: 'Дуже захоплено'
            },
            {
              value: '🤔',
              label: 'Подумки'
            },
            {
              value: '🙄',
              label: 'Зневажливий погляд'
            },
            {
              value: '😜',
              label: 'Лукавий'
            },
            {
              value: '😲',
              label: 'Застогнав'
            },
            {
              value: '😭',
              label: 'Плачу до сліз'
            },
            {
              value: '🤯',
              label: 'Вибух мозку'
            },
            {
              value: '😰',
              label: 'Холодний пот'
            },
            {
              value: '😱',
              label: 'Відлякало'
            },
            {
              value: '🤪',
              label: 'Безлад'
            },
            {
              value: '😵',
              label: 'Обморок'
            },
            {
              value: '😡',
              label: 'Гнів'
            },
            {
              value: '🥳',
              label: 'Привітання'
            },
            {
              value: '🤡',
              label: 'Як жартую, я ж це я'
            },
            {
              value: '🤫',
              label: 'Ш-ш-ш'
            },
            {
              value: '🐒',
              label: 'Мавпа'
            },
            {
              value: '🤭',
              label: 'Посмішка без слів'
            },
            {
              value: '🐂',
              label: 'Бик'
            },
            {
              value: '🍺',
              label: 'Пиво'
            }
          ]
  
  })
  
  el.onSelect(l => {
  document.querySelector(".emoji-picker").value+=l
  console.log(value);
  })