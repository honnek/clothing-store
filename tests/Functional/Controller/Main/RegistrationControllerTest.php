<?php

namespace App\Tests\Functional\Controller\Main;

use App\Entity\User;
use App\Repository\UserRepository;
use App\Tests\TestUtils\Fixtures\UserFixtures;
use Symfony\Bundle\FrameworkBundle\Test\WebTestCase;

/**
 * @group functional
 */
class RegistrationControllerTest extends WebTestCase
{
    public function testEmailSendAndCreateUser(): void
    {
        $newUserEmail = 'newTestUser1@test.com';
        $newUserPassword = 'testUser123';

        $client = static::createClient();
        $client->request("GET", '/en/registration');
        $crawler = $client->submitForm('SIGN UP', [
            'registration_form[email]' => $newUserEmail,
            'registration_form[plainPassword]' => $newUserPassword,
            'registration_form[agreeTerms]' => true,
        ]);

        $this->assertEmailCount(
            1,
        );

        /** @var User $user */
        $user = static::getContainer()->get(UserRepository::class)->findOneBy(['email' => $newUserEmail]);

        $this->assertNotNull($user);
        $this->assertSame($newUserEmail, $user->getEmail());
    }

    public function testEmailDuplicateError(): void
    {
        $client = static::createClient();

        $duplicateEmail = UserFixtures::USER_1_EMAIL;
        $password = 'user7897';

        $client->request("GET", '/en/registration');
        $crawler = $client->submitForm('SIGN UP', [
            'registration_form[email]' => $duplicateEmail,
            'registration_form[plainPassword]' => $password,
            'registration_form[agreeTerms]' => true,
        ]);

        $this->assertResponseIsSuccessful();
        $this->assertSelectorTextContains('div', 'There is already an account with this email');
    }

    public function testShortPasswordError(): void
    {
        $client = static::createClient();

        $email = 'user7852@test.com';
        $password = '123';

        $client->request("GET", '/en/registration');
        $crawler = $client->submitForm('SIGN UP', [
            'registration_form[email]' => $email,
            'registration_form[plainPassword]' => $password,
            'registration_form[agreeTerms]' => true,
        ]);

        $this->assertResponseIsSuccessful();
        $this->assertSelectorTextContains('div', 'Your password should be at least 6 characters');
    }
}
