<?php

declare(strict_types=1);

namespace DoctrineMigrations;

use Doctrine\DBAL\Schema\Schema;
use Doctrine\Migrations\AbstractMigration;

/**
 * Auto-generated Migration: Please modify to your needs!
 */
final class Version20231214194703 extends AbstractMigration
{
    public function getDescription(): string
    {
        return '';
    }

    public function up(Schema $schema): void
    {
        // this up() migration is auto-generated, please modify it to your needs
        $this->addSql('DROP SEQUENCE test_table_id_seq CASCADE');
        $this->addSql('DROP SEQUENCE test_api_entity_id_seq CASCADE');
        $this->addSql('DROP TABLE test_table');
        $this->addSql('DROP TABLE test_api_entity');
        $this->addSql('CREATE TABLE houses (zip_code INT, price VARCHAR(50), beds INT, 
        baths INT, living_space INT, address TEXT, city VARCHAR(75), state VARCHAR(75), zip_code_population FLOAT, 
        zip_code_density FLOAT, county VARCHAR(75), median_household_income INT, latitude FLOAT, longitude NUMERIC)');
    }

    public function down(Schema $schema): void
    {
        // this down() migration is auto-generated, please modify it to your needs
        $this->addSql('CREATE SEQUENCE test_table_id_seq INCREMENT BY 1 MINVALUE 1 START 1');
        $this->addSql('CREATE SEQUENCE test_api_entity_id_seq INCREMENT BY 1 MINVALUE 1 START 1');
        $this->addSql('CREATE TABLE test_table (id SERIAL NOT NULL, name VARCHAR(255) NOT NULL, PRIMARY KEY(id))');
        $this->addSql('CREATE TABLE test_api_entity (id INT NOT NULL, title VARCHAR(255) NOT NULL, description VARCHAR(255) NOT NULL, is_deleted BOOLEAN NOT NULL, PRIMARY KEY(id))');
        $this->addSql('DROP TABLE houses');
    }
}
