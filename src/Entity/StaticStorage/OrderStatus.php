<?php

namespace App\Entity\StaticStorage;

/**
 * Класс для работы со статусами заказа
 */
class OrderStatus
{
    /** Статус - заказ создан */
    public const CREATED = 0;

    /** Статус - заказ в обработке */
    public const PROCESSED = 1;

    /** Статус - заказ укомплектован */
    public const COMPLECTED = 2;

    /** Статус - заказ доставляется */
    public const DELIVERED = 3;

    /** Статус - заказ отменен */
    public const DENIED = 4;


    /**
     * Возвращает список статусов
     * @return array
     */
    public static function getList(): array
    {
        return [
            self::CREATED => 'Created',
            self::PROCESSED => 'Processed',
            self::COMPLECTED => 'Complected',
            self::DELIVERED => 'Delivered',
            self::DENIED => 'Denied',
        ];
    }
}
