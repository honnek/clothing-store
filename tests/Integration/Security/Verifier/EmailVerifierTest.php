<?php

namespace App\Tests\Integration\Security\Verifier;

use App\Entity\User;
use App\Repository\CategoryRepository;
use App\Repository\OrderRepository;
use App\Repository\ProductRepository;
use App\Repository\UserRepository;
use App\Security\Verifier\EmailVerifier;
use App\Tests\TestUtils\Fixtures\UserFixtures;
use Symfony\Bundle\FrameworkBundle\Test\KernelTestCase;

/**
 * @group integration
 */
class EmailVerifierTest extends KernelTestCase
{
    /**
     * @var object|UserRepository
     */
    private $userRepository;

    /**
     * @var object|EmailVerifier
     */
    private $emailVerifier;

    public function setUp(): void
    {
        $kernel = self::bootKernel();

        $this->emailVerifier = static::getContainer()->get(EmailVerifier::class);
        $this->userRepository = static::getContainer()->get(UserRepository::class);
    }

    public function testGenerateEmailSignature(): void
    {
        $currentTime = new \DateTimeImmutable();
        $user = $this->userRepository->findOneBy(['email' => UserFixtures::USER_1_EMAIL]);
        $user->setIsVerified(false);

        $emailSignature = $this->emailVerifier->generateEmailSignature('main_verify_email', $user);
        $this->assertGreaterThan($currentTime, $emailSignature->getExpiresAt());
    }

    public function testHandleEmailConfirmation(): void
    {
        $currentTime = new \DateTimeImmutable();
        $user = $this->userRepository->findOneBy(['email' => UserFixtures::USER_1_EMAIL]);
        $user->setIsVerified(false);

        $emailSignature = $this->emailVerifier->generateEmailSignature('main_verify_email', $user);
        $this->emailVerifier->handleEmailConfirmation($emailSignature->getSignedUrl(), $user);

        $this->assertTrue($user->isVerified());
    }
}
