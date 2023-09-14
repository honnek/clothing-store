<?php

namespace App\Tests\Integration\Security\Verifier;

use App\Repository\CategoryRepository;
use App\Repository\OrderRepository;
use App\Repository\ProductRepository;
use App\Repository\UserRepository;
use Symfony\Bundle\FrameworkBundle\Test\KernelTestCase;

/**
 * @group integration
 */
class EmailVerifierTest extends KernelTestCase
{
    public function setUp(): void
    {
        $kernel = self::bootKernel();

        /** @var UserRepository $userRep */
        $userRep = static::getContainer()->get(UserRepository::class);
        $categoryRep = static::getContainer()->get(CategoryRepository::class);
        $productRep = static::getContainer()->get(ProductRepository::class);
        $orderRep = static::getContainer()->get(OrderRepository::class);
        $userRep = static::getContainer()->get(UserRepository::class);

//        var_dump($userRep->findAll());
    }

    public function testSomething(): void
    {
        $kernel = self::bootKernel();

        $this->assertSame('test', $kernel->getEnvironment());
        // $routerService = static::getContainer()->get('router');
        // $myCustomService = static::getContainer()->get(CustomService::class);
    }
}
