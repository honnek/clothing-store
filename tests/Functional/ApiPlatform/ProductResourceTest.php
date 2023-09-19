<?php

namespace App\Tests\Functional\ApiPlatform;

use App\Entity\Product;
use App\Entity\User;
use App\Repository\ProductRepository;
use App\Repository\UserRepository;
use App\Tests\TestUtils\Fixtures\UserFixtures;
use Symfony\Bundle\FrameworkBundle\Test\WebTestCase;
use Symfony\Component\HttpFoundation\Response;

/**
 * @group functional
 */
class ProductResourceTest extends WebTestCase
{
    private const REQUEST_HEADERS = [
        'HTTP_ACCEPT' => 'application/ld+json',
        'CONTENT_TYPE' => 'application/json'
    ];

    private const REQUEST_HEADERS_FOR_PATCH = [
        'HTTP_ACCEPT' => 'application/ld+json',
        'CONTENT_TYPE' => 'application/ld+json'
    ];

    /**
     * @var string
     */
    protected string $uriKey = '/api/products';

    public function testGetProducts(): void
    {
        $client = static::createClient();

        $client->request('GET', $this->uriKey, [], [], self::REQUEST_HEADERS);

        $this->assertResponseStatusCodeSame(Response::HTTP_OK);
    }

    public function testGetProduct(): void
    {
        $client = static::createClient();
        /** @var Product $product - любой продукт из фикстур */
        $product = static::getContainer()->get(ProductRepository::class)->findOneBy([]);

        $uri = $this->uriKey . '/' . $product->getUuid();
        $client->request('GET', $uri, [], [], self::REQUEST_HEADERS);

        $this->assertResponseStatusCodeSame(Response::HTTP_OK);
    }

    public function testAdminCreateProduct(): void
    {
        $client = static::createClient();
        /** @var User $user */
        $user = static::getContainer()->get(UserRepository::class)
            ->findOneBy(['email' => UserFixtures::USER_ADMIN_1_EMAIL]);

        $client->loginUser($user);

        $context = [
            'title' => 'New Product',
            'price' => '150',
            'quality' => 5
        ];

        $client->request('POST', $this->uriKey, [], [], self::REQUEST_HEADERS, json_encode($context));

        $this->assertResponseStatusCodeSame(Response::HTTP_CREATED);
    }

    public function testUserCreateProduct(): void
    {
        $client = static::createClient();
        /** @var User $user */
        $user = static::getContainer()->get(UserRepository::class)
            ->findOneBy(['email' => UserFixtures::USER_1_EMAIL]);

        $client->loginUser($user);

        $context = [
            'title' => 'New Product',
            'price' => '150',
            'quality' => 5
        ];

        $client->request('POST', $this->uriKey, [], [], self::REQUEST_HEADERS, json_encode($context));

        $this->assertResponseStatusCodeSame(Response::HTTP_FORBIDDEN);
    }

    public function testPatchProduct(): void
    {
        $client = static::createClient();
        /** @var User $user */
        $user = static::getContainer()->get(UserRepository::class)
            ->findOneBy(['email' => UserFixtures::USER_ADMIN_1_EMAIL]);

        /** @var Product $product - любой продукт из фикстур */
        $product = static::getContainer()->get(ProductRepository::class)->findOneBy([]);
        $uri = $this->uriKey . '/' . $product->getUuid();

        $client->request('PATCH', $uri, [], [], self::REQUEST_HEADERS);

        $client->loginUser($user);

        $context = [
            'title' => 'Updated Product',
            'price' => '150',
            'quality' => 5
        ];

        $client->request('PATCH', $uri, [], [], self::REQUEST_HEADERS_FOR_PATCH, json_encode($context));

        $this->assertResponseStatusCodeSame(Response::HTTP_OK);
    }
}
