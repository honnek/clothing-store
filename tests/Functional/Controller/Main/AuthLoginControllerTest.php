<?php

namespace App\Tests\Functional\Controller\Main;

use App\Tests\TestUtils\Fixtures\UserFixtures;
use Symfony\Component\HttpFoundation\Response;
use Symfony\Component\Panther\Client;
use Symfony\Component\Panther\PantherTestCase;

/**
 * @group functional
 */
class AuthLoginControllerTest extends PantherTestCase
{

    public function testLogin(): void
    {
        $client = static::createClient();

        $client->request('GET', '/en/login');
        $crawler = $client->submitForm('LOG IN', [
            'email' => UserFixtures::USER_1_EMAIL,
            'password' => UserFixtures::USER_1_PASSWORD,
        ]);

        $this->assertResponseRedirects('/en/profile', Response::HTTP_FOUND);

        $client->followRedirect();

        $this->assertResponseIsSuccessful();
    }

    public function testLoginWithPantherClient(): void
    {
        $client = self::createPantherClient(['browser' => self::CHROME]);

        $client->request('GET', '/en/login');
        $client->submitForm('LOG IN', [
            'email' => UserFixtures::USER_1_EMAIL,
            'password' => UserFixtures::USER_1_PASSWORD,
        ]);

        $this->assertSame(self::$baseUri . '/en/profile', $client->getCurrentURL());

        $this->assertPageTitleContains('My profile - RankedChoice');
        $this->assertSelectorTextContains('#page_header_title', 'Welcome, to your profile!');
    }

//    public function testLoginWithSeleniumClient(): void
//    {
//        self::createPantherClient();
//        self::startWebServer();
//
//        $client = Client::createSeleniumClient(
//            'http://selenium-hub:4444/wd/hub',
//            null,
//            'http://ranked-choiceshop-chrome-1:5900',
//        );
//
//        $client->request('GET', '/en/login');
//        $crawler = $client->submitForm('LOG IN', [
//            'email' => UserFixtures::USER_1_EMAIL,
//            'password' => UserFixtures::USER_1_PASSWORD,
//        ]);
//
//        $client->takeScreenshot('var/tests-screenshots/test.png');
//
//        $this->assertSame($crawler->filter('#page_header_title')->text(), 'Welcome, to your profile!');
//    }
}
