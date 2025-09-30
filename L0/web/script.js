document.addEventListener('DOMContentLoaded', () => {
    const searchForm = document.getElementById('search-form');
    const orderIdInput = document.getElementById('order-id-input');
    const resultDisplay = document.getElementById('result-display');
    const errorDisplay = document.getElementById('error-display');
    const submitButton = searchForm.querySelector('button[type="submit"]');

    const buttonText = submitButton ? submitButton.querySelector('.button-text') : null;
    const spinner = submitButton ? submitButton.querySelector('.spinner') : null;

    if (!buttonText || !spinner) {
        console.warn('Внимание: не найдены элементы button-text или spinner в кнопке');
    }

    searchForm.addEventListener('submit', async (event) => {
        event.preventDefault();

        hideResults();

        const orderId = orderIdInput.value.trim();

        if (!orderId) {
            showError({
                title: "Ошибка ввода",
                message: "Пожалуйста, введите ID заказа."
            });
            return;
        }

        const url = `/order/${encodeURIComponent(orderId)}`;

        try {
            setLoading(true);

            const controller = new AbortController();
            const timeoutId = setTimeout(() => controller.abort(), 15000);

            const startTime = Date.now();

            const response = await fetch(url, {
                method: 'GET',
                signal: controller.signal,
                headers: {
                    'Accept': 'application/json',
                    'Content-Type': 'application/json'
                }
            });

            clearTimeout(timeoutId);

            const requestTime = Date.now() - startTime;
            console.log(`Response received in ${requestTime}ms`);
            console.log('Response status:', response.status);
            console.log('Response ok:', response.ok);

            const slowResponse = requestTime > 10000;

            if (response.status === 404) {
                let errorMessage = `Заказ с ID "${orderId}" не найден`;

                try {
                    const contentType = response.headers.get("content-type");
                    if (contentType && contentType.includes("application/json")) {
                        const errorData = await response.json();
                        errorMessage = errorData.message || errorData.error || errorMessage;
                    } else {
                        const textError = await response.text();
                        if (textError && !textError.includes('<html')) {
                            const cleanError = textError.replace(/\\n/g, '').trim();
                            if (cleanError && cleanError !== 'Order not found') {
                                errorMessage = cleanError;
                            }
                        }
                    }
                } catch (e) {
                    console.log('Не удалось прочитать тело ошибки 404:', e);
                }

                if (slowResponse) {
                    errorMessage += ` (Внимание: сервер отвечал ${Math.round(requestTime/1000)} секунд)`;
                }

                showError({
                    title: "Заказ не найден",
                    message: errorMessage,
                    code: 404
                });
                return;
            }

            const contentType = response.headers.get("content-type");
            const isJson = contentType && contentType.includes("application/json");

            if (!response.ok) {
                let errorData = {
                    message: `Ошибка ${response.status}`,
                    status: response.status
                };

                if (isJson) {
                    try {
                        const jsonError = await response.json();
                        console.log('Error from server:', jsonError);
                        errorData.message = jsonError.message ||
                            jsonError.error ||
                            jsonError.detail ||
                            jsonError.msg ||
                            `Ошибка ${response.status}`;
                    } catch (e) {
                        console.error("Не удалось распарсить JSON ошибку:", e);
                    }
                } else {
                    try {
                        const textError = await response.text();
                        if (textError && !textError.includes('<html')) {
                            errorData.message = textError || `Ошибка ${response.status}`;
                        }
                    } catch (e) {
                        console.error("Не удалось прочитать текст ошибки:", e);
                    }
                }

                const errorTitle = getErrorTitleByStatus(response.status);

                if (slowResponse) {
                    errorData.message += ` (Время ответа: ${Math.round(requestTime/1000)} сек)`;
                }

                showError({
                    title: errorTitle,
                    message: errorData.message,
                    code: response.status
                });
                return;
            }

            if (!isJson) {
                throw new Error('Сервер вернул некорректный формат данных (ожидался JSON)');
            }

            try {
                const data = await response.json();
                console.log('Success data:', data);

                if (slowResponse) {
                    console.warn(`Внимание: сервер ответил за ${Math.round(requestTime/1000)} секунд`);
                }

                showResult(data);
            } catch (parseError) {
                console.error('Ошибка парсинга JSON:', parseError);
                throw new Error('Не удалось обработать данные от сервера');
            }

        } catch (error) {
            console.error('Catch block error:', error);

            if (error.name === 'AbortError') {
                showError({
                    title: "Превышено время ожидания",
                    message: `Сервер не ответил за 15 секунд. Возможно, проблемы с соединением или ID заказа "${orderId}" не существует. Попробуйте позже.`
                });
            } else {
                handleFetchError(error);
            }
        } finally {
            setLoading(false);
        }
    });


    function showResult(data) {
        if (!data || !data.order_uid) {
            showError({
                title: "Ошибка данных",
                message: "Получены неполные данные о заказе"
            });
            return;
        }

        let formattedDate = 'Не указана';
        if (data.date_created) {
            try {
                formattedDate = new Date(data.date_created).toLocaleString('ru-RU', {
                    year: 'numeric',
                    month: 'long',
                    day: 'numeric',
                    hour: '2-digit',
                    minute: '2-digit'
                });
            } catch (e) {
                console.error('Ошибка форматирования даты:', e);
                formattedDate = data.date_created;
            }
        }

        const orderHtml = `
            <div class="order-card">
                <h2>Информация о заказе</h2>
                <div class="order-details">
                    <p><strong>ID Заказа:</strong> <span>${escapeHtml(data.order_uid)}</span></p>
                    <p><strong>Дата создания:</strong> <span>${formattedDate}</span></p>
                    <p><strong>Служба доставки:</strong> <span>${escapeHtml(data.delivery_service || 'Не указана')}</span></p>
                </div>

                ${data.items && data.items.length > 0 ? `
                    <h2>Состав заказа</h2>
                    <table class="items-table">
                        <thead>
                            <tr>
                                <th>Название</th>
                                <th>Бренд</th>
                                <th>Цена</th>
                                <th>Скидка</th>
                                <th>Итого</th>
                            </tr>
                        </thead>
                        <tbody>
                            ${data.items.map(item => `
                                <tr>
                                    <td>${escapeHtml(item.name)} (Арт: ${escapeHtml(item.nm_id)})</td>
                                    <td>${escapeHtml(item.brand)}</td>
                                    <td>${item.price} ${escapeHtml(data.currency)}</td>
                                    <td>${item.sale}%</td>
                                    <td><strong>${item.total_price} ${escapeHtml(data.currency)}</strong></td>
                                </tr>
                            `).join('')}
                        </tbody>
                    </table>
                ` : '<p>Информация о товарах отсутствует</p>'}

                <h2>Финансовая информация</h2>
                <div class="order-summary">
                     <p><strong>Сумма товаров:</strong> <span>${data.goods_total || 0} ${escapeHtml(data.currency)}</span></p>
                     <p><strong>Стоимость доставки:</strong> <span>${data.delivery_cost || 0} ${escapeHtml(data.currency)}</span></p>
                     <hr>
                     <p class="total"><strong>Общая сумма заказа:</strong> <span>${data.amount || 0} ${escapeHtml(data.currency)}</span></p>
                </div>
            </div>
        `;

        resultDisplay.innerHTML = orderHtml;
        resultDisplay.classList.remove('hidden');
    }

    function showError({ title = "Ошибка", message, code }) {
        const errorHtml = `
            <div class="error-card">
                <div class="error-header">
                    <span class="error-title">${escapeHtml(title)}</span>
                    ${code ? `<span class="error-code">Код: ${code}</span>` : ''}
                </div>
                <div class="error-body">
                    ${escapeHtml(message)}
                </div>
            </div>
        `;
        errorDisplay.innerHTML = errorHtml;
        errorDisplay.classList.remove('hidden');
    }

    function setLoading(isLoading) {
        if (!submitButton) {
            console.error('Кнопка отправки не найдена');
            return;
        }

        submitButton.disabled = isLoading;

        if (buttonText && spinner) {
            if (isLoading) {
                buttonText.classList.add('hidden');
                spinner.classList.remove('hidden');
            } else {
                buttonText.classList.remove('hidden');
                spinner.classList.add('hidden');
            }
        }
    }

    function hideResults() {
        resultDisplay.classList.add('hidden');
        errorDisplay.classList.add('hidden');
        resultDisplay.innerHTML = "";
        errorDisplay.innerHTML = "";
    }

    function handleFetchError(error) {
        if (error instanceof TypeError) {
            if (error.message.includes('Failed to fetch')) {
                showError({
                    title: "Ошибка соединения",
                    message: "Не удалось подключиться к серверу. Проверьте интернет-соединение."
                });
            } else if (error.message.includes('Load failed')) {
                showError({
                    title: "Ошибка загрузки",
                    message: "Не удалось загрузить данные. Возможно, сервер недоступен."
                });
            } else {
                showError({
                    title: "Ошибка сети",
                    message: error.message || "Проблема с сетевым соединением"
                });
            }
        } else if (error instanceof SyntaxError) {
            showError({
                title: "Ошибка данных",
                message: "Получены некорректные данные от сервера"
            });
        } else {
            showError({
                title: "Произошла ошибка",
                message: error.message || "Неизвестная ошибка. Попробуйте позже."
            });
        }
    }

    function getErrorTitleByStatus(status) {
        const statusTitles = {
            400: "Неверный запрос",
            404: "Не найдено",
            500: "Ошибка сервера",
        };
        return statusTitles[status] || "Ошибка сервера";
    }

    function escapeHtml(text) {
        if (!text) return '';
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }
});