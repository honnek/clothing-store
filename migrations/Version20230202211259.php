<?php

declare(strict_types=1);

namespace DoctrineMigrations;

use Doctrine\DBAL\Schema\Schema;
use Doctrine\Migrations\AbstractMigration;

/**
 * Миграция для изменения на false поля isDeleted в таблице User
 */
final class Version20230202211259 extends AbstractMigration
{
    public function getDescription(): string
    {
        return '';
    }

    public function up(Schema $schema): void
    {
        $this->addSql('UPDATE "user" SET is_deleted=\'0\' WHERE is_deleted IS NULL');

    }

    public function down(Schema $schema): void
    {
        $this->addSql('UPDATE "user" SET is_deleted=NULL WHERE is_deleted IS NOT NULL');
    }
}
