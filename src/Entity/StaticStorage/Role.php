<?php

namespace App\Entity\StaticStorage;

class Role
{

    /** Роль - обычный пользователь */
    public const USER = 'ROLE_USER';

    /** Роль - Админ */
    public const ADMIN = 'ROLE_ADMIN';

    /** Роль - Супер Админ */
    public const SUPER_ADMIN = 'ROLE_SUPER_ADMIN';

    /**
     * @return string[]
     */
    public static function getList(): array
    {
        return [
            self::USER => 'ROLE_USER',
            self::ADMIN => 'ROLE_ADMIN',
            self::SUPER_ADMIN => 'ROLE_SUPER_ADMIN',
        ];
    }
}
