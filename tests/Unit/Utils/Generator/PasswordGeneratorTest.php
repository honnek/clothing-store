<?php

namespace App\Tests\Unit\Utils\Generator;

use App\Utils\Generator\PasswordGenerator;
use PHPUnit\Framework\TestCase;

/**
 * @group unit
 */
class PasswordGeneratorTest extends TestCase
{
    public function testGeneratePassword(): void
    {
        $length = 8;
        $password = PasswordGenerator::generatePassword($length);

        self::assertSame($length, strlen($password));
    }
}
