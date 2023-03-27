<?php

namespace App\Entity;

use ApiPlatform\Metadata\ApiResource;
use ApiPlatform\Metadata\Delete;
use ApiPlatform\Metadata\Get;
use ApiPlatform\Metadata\GetCollection;
use ApiPlatform\Metadata\Post;
use App\Repository\ProductImageRepository;
use Doctrine\ORM\Mapping as ORM;
use Symfony\Component\Serializer\Annotation\Groups;

#[ORM\Entity(repositoryClass: ProductImageRepository::class)]
#[ApiResource(
    operations: [
        new GetCollection(normalizationContext: ['groups' => ['product_image:list']]),
        new Delete(formats: [ 'jsonld', 'json']),
        new Get(normalizationContext: ['groups' => ['product_image:item']]),
    ]
)]
class ProductImage
{
    #[ORM\Id]
    #[ORM\GeneratedValue]
    #[ORM\Column]
    #[Groups(["cart:item", "cart:list", "cart_product:item", "cart_product:list"])]
    private ?int $id = null;

    #[ORM\ManyToOne(inversedBy: 'productImages')]
    #[ORM\JoinColumn(nullable: false)]
    private ?Product $product = null;

    #[ORM\Column(length: 255)]
    private ?string $filenameBig = null;

    #[ORM\Column(length: 255)]
    private ?string $filenameMiddle = null;

    #[ORM\Column(length: 255)]
    #[Groups(["cart:item", "cart:list", "cart_product:item", "cart_product:list"])]
    private ?string $filenameSmall = null;

    public function getId(): ?int
    {
        return $this->id;
    }

    public function getProduct(): ?Product
    {
        return $this->product;
    }

    public function setProduct(?Product $product): self
    {
        $this->product = $product;

        return $this;
    }

    public function getFilenameBig(): ?string
    {
        return $this->filenameBig;
    }

    public function setFilenameBig(string $filenameBig): self
    {
        $this->filenameBig = $filenameBig;

        return $this;
    }

    public function getFilenameMiddle(): ?string
    {
        return $this->filenameMiddle;
    }

    public function setFilenameMiddle(string $filenameMiddle): self
    {
        $this->filenameMiddle = $filenameMiddle;

        return $this;
    }

    public function getFilenameSmall(): ?string
    {
        return $this->filenameSmall;
    }

    public function setFilenameSmall(string $filenameSmall): self
    {
        $this->filenameSmall = $filenameSmall;

        return $this;
    }
}
