<?php

namespace App\Form\Handler;

use App\Entity\Category;
use App\Form\DTO\EditCategoryModel;
use App\Utils\Manager\CategoryManager;
use Symfony\Component\String\Slugger\SluggerInterface;

class CategoryFormHandler
{

    private SluggerInterface $slugger;
    private CategoryManager $categoryManager;

    public function __construct(CategoryManager $categoryManager, SluggerInterface $slugger)
    {
        $this->categoryManager = $categoryManager;
        $this->slugger = $slugger;
    }

    /**
     * @param EditCategoryModel $editCategoryModel
     * @return Category
     */
    public function processEditForm(EditCategoryModel $editCategoryModel): Category
    {
        $category = new Category();

        if ($editCategoryModel->id) {
            $category = $this->categoryManager->getRepository()->find($editCategoryModel->id);
        }

        $category->setTitle($editCategoryModel->title);
        $category->setSlug(strtolower($this->slugger->slug($editCategoryModel->title)));
        $this->categoryManager->save($category);

        return $category;
    }
}
