<?php

namespace App\Tests\Integration\Security\Verifier;

use App\Entity\User;
use App\Repository\UserRepository;
use App\Security\Verifier\EmailVerifier;
use App\Tests\TestUtils\Fixtures\UserFixtures;
use Symfony\Bundle\FrameworkBundle\Test\KernelTestCase;
use SymfonyCasts\Bundle\VerifyEmail\Model\VerifyEmailSignatureComponents;

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

    public function testGenerateEmailSignatureAndConfirmation(): void
    {
        $user = $this->userRepository->findOneBy(['email' => UserFixtures::USER_1_EMAIL]);
        $user->setIsVerified(false);

        $emailSignature = $this->checkGenerateEmailSignature($user);

        $this->checkEmailConfirmation($emailSignature, $user);
    }

    private function checkGenerateEmailSignature(User $user): VerifyEmailSignatureComponents
    {
        $currentTime = new \DateTimeImmutable();
        $emailSignature = $this->emailVerifier->generateEmailSignature('main_verify_email', $user);

        $this->assertGreaterThan($currentTime, $emailSignature->getExpiresAt());

        return $emailSignature;
    }

    private function checkEmailConfirmation(VerifyEmailSignatureComponents $signatureComponents, User $user): void
    {
        $this->assertFalse($user->isVerified());

        $this->emailVerifier->handleEmailConfirmation($signatureComponents->getSignedUrl(), $user);
        $this->assertTrue($user->isVerified());
    }

//    public function testGenerateEmailSignature(): void
//    {
//        $currentTime = new \DateTimeImmutable();
//        $user = $this->userRepository->findOneBy(['email' => UserFixtures::USER_1_EMAIL]);
//        $user->setIsVerified(false);
//
//        $emailSignature = $this->emailVerifier->generateEmailSignature('main_verify_email', $user);
//        $this->assertGreaterThan($currentTime, $emailSignature->getExpiresAt());
//    }
//
//    public function testHandleEmailConfirmation(): void
//    {
//        $user = $this->userRepository->findOneBy(['email' => UserFixtures::USER_1_EMAIL]);
//        $user->setIsVerified(false);
//
//        $emailSignature = $this->emailVerifier->generateEmailSignature('main_verify_email', $user);
//        $this->emailVerifier->handleEmailConfirmation($emailSignature->getSignedUrl(), $user);
//
//        $this->assertTrue($user->isVerified());
//    }
}
